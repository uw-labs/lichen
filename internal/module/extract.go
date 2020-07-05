package module

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/uw-labs/lichen/internal/model"
	"github.com/uw-labs/lichen/internal/module/buildinfo"
)

func Extract(ctx context.Context, paths ...string) ([]model.BuildInfo, error) {
	output, err := goVersion(ctx, paths)
	if err != nil {
		return nil, err
	}

	parsed, err := buildinfo.Parse(output)
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
