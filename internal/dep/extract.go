package dep

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/utilitywarehouse/lichen/internal/model"
)

func Extract(ctx context.Context, paths ...string) ([]model.Binary, error) {
	output, err := goVersion(ctx, paths)
	if err != nil {
		return nil, err
	}

	parsed, err := parseOutput(output)
	if err != nil {
		return nil, err
	}
	if len(parsed) == 0 {
		return nil, fmt.Errorf("could not extract module information from binaries: %v", paths)
	}
	return parsed, nil
}

func goVersion(ctx context.Context, paths []string) (string, error) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return "", err
	}

	tempDir, err := ioutil.TempDir("", "lichen")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.Remove(tempDir)

	args := []string{"version", "-m"}
	args = append(args, paths...)

	cmd := exec.CommandContext(ctx, goBin, args...)
	cmd.Dir = tempDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), err
}

func parseOutput(output string) ([]model.Binary, error) {
	var (
		lines   = strings.Split(output, "\n")
		results = make([]model.Binary, 0)
		current model.Binary
	)
	for _, l := range lines {
		parts := strings.Fields(l)
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "path":
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid path: %s", l)
			}
			if current.Path != "" {
				results = append(results, current)
			}
			current = model.Binary{Path: parts[1]}
		case "mod":
		case "dep":
			if len(parts) < 3 {
				return nil, fmt.Errorf("invalid module: %s", l)
			}
			current.Refs = append(current.Refs, model.Reference{
				Path:    parts[1],
				Version: parts[2],
			})
		default:
			continue
		}
	}
	if current.Path != "" {
		results = append(results, current)
	}
	return results, nil
}
