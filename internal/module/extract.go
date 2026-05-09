package module

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-multierror"
	"github.com/uw-labs/lichen/internal/buildinfo"
	"github.com/uw-labs/lichen/internal/model"
)

// Extract extracts build information from the supplied binaries
func Extract(ctx context.Context, paths ...string) ([]model.BuildInfo, error) {
	output, err := goVersion(ctx, paths)
	if err != nil {
		return nil, err
	}

	parsed, err := buildinfo.Parse(output)
	if err != nil {
		return nil, err
	}
	if err := verifyExtracted(parsed, paths); err != nil {
		return nil, fmt.Errorf("could not extract module information: %w", err)
	}
	return parsed, nil
}

// verifyExtracted ensures all paths requests are covered by the parsed output
func verifyExtracted(extracted []model.BuildInfo, requested []string) (err error) {
	buildInfos := make(map[string]struct{}, len(extracted))
	for _, binary := range extracted {
		buildInfos[binary.Path] = struct{}{}
	}
	for _, path := range requested {
		if _, found := buildInfos[path]; !found {
			err = multierror.Append(err, fmt.Errorf("modules could not be obtained from %[1]s (hint: run `go version -m %[1]q`)", path))
		}
	}
	return
}

// goVersion runs `go version -m [paths ...]` and returns the output
func goVersion(ctx context.Context, paths []string) (string, error) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return "", err
	}

	tempDir, err := os.MkdirTemp("", "lichen")
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
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("error when running 'go version': %w - stderr: %s", err, exitErr.Stderr)
		}
		return "", fmt.Errorf("error when running 'go version': %w", err)
	}

	return string(out), err
}
