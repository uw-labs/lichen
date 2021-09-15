package module

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/uw-labs/lichen/internal/model"
)

func Fetch(ctx context.Context, refs []model.ModuleReference) ([]model.Module, error) {
	if len(refs) == 0 {
		return []model.Module{}, nil
	}

	var (
		f   fetcher
		err error
	)

	if f.goBin, err = exec.LookPath("go"); err != nil {
		return nil, err
	}

	if f.tempDir, err = ioutil.TempDir("", "lichen"); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.Remove(f.tempDir)

	if strings.Contains(os.Getenv("GOFLAGS"), "-mod=vendor") {
		f.vendorMode = true
	} else if _, err := os.Stat("./vendor/modules.txt"); err == nil {
		f.vendorMode = true
	}

	args := []string{"mod", "download", "-json"}
	if f.vendorMode {
		args = []string{"list", "-m", "-json", "-mod=readonly"}
	}

	for _, ref := range refs {
		if !ref.IsLocal() {
			args = append(args, ref.String())
		}
	}

	if err = f.fetch(ctx, args); err != nil {
		return nil, err
	}

	// add local modules, as these won't be included in the set returned by `go mod download`
	for _, ref := range refs {
		if ref.IsLocal() {
			f.modules = append(f.modules, model.Module{
				ModuleReference: ref,
			})
		}
	}

	// sanity check: all modules should have been covered in the output from `go mod download`
	if err := f.verifyFetched(refs); err != nil {
		return nil, fmt.Errorf("failed to fetch all modules: %w", err)
	}

	return f.modules, nil
}

type fetcher struct {
	goBin      string
	tempDir    string
	vendorMode bool

	modules []model.Module
}

func (f *fetcher) fetch(ctx context.Context, args []string) error {
	cmd := exec.CommandContext(ctx, f.goBin, args...)
	cmd.Dir = f.tempDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to fetch: %w (output: %s)", err, string(out))
	}

	// parse JSON output from `go mod list` or `go mod download`
	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var m model.Module
		if err := dec.Decode(&m); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		if f.vendorMode {
			if _, err := os.Stat("./vendor/" + m.Path); err == nil {
				m.Dir = "./vendor/" + m.Path
			}
		}

		if m.Dir == "" {
			continue
		}

		f.modules = append(f.modules, m)
	}

	return nil
}

func (f *fetcher) verifyFetched(requested []model.ModuleReference) (err error) {
	fetchedRefs := make(map[model.ModuleReference]struct{}, len(f.modules))
	for _, module := range f.modules {
		fetchedRefs[module.ModuleReference] = struct{}{}
	}
	for _, ref := range requested {
		if _, found := fetchedRefs[ref]; !found {
			err = multierror.Append(err, fmt.Errorf("module %s could not be resolved", ref))
		}
	}
	return
}
