package dep

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"github.com/utilitywarehouse/golly/internal/model"
)

func Fetch(ctx context.Context, refs []model.Reference) ([]model.Module, error) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return nil, err
	}

	args := []string{"mod", "download", "-json"}
	for _, mod := range refs {
		args = append(args, fmt.Sprintf("%s@%s", mod.Path, mod.Version))
	}

	cmd := exec.CommandContext(ctx, goBin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w (output: %s)", err, string(out))
	}

	modules := make([]model.Module, 0)
	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var m model.Module
		if err := dec.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		modules = append(modules, m)
	}

	return modules, nil
}
