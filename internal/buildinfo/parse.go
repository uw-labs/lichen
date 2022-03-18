package buildinfo

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/uw-labs/lichen/internal/model"
)

var goVersionRgx = regexp.MustCompile(`^(.*?): (?:(?:devel )?go[0-9]+|devel \+[0-9a-f]+)`)

// Parse parses build info details as returned by `go version -m [bin ...]`
func Parse(info string) ([]model.BuildInfo, error) {
	var (
		lines       = strings.Split(info, "\n")
		results     = make([]model.BuildInfo, 0)
		current     model.BuildInfo
		replacement bool
	)
	for _, l := range lines {
		// ignore blank lines
		if l == "" {
			continue
		}

		// start of new build info output
		if !strings.HasPrefix(l, "\t") {
			matches := goVersionRgx.FindStringSubmatch(l)
			if len(matches) != 2 {
				return nil, fmt.Errorf("unrecognised version line: %s", l)
			}
			if current.Path != "" {
				results = append(results, current)
			}
			current = model.BuildInfo{Path: matches[1]}
			continue
		}

		// inside build info output
		parts := strings.Split(l, "\t")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid build info line: %s", l)
		}
		if replacement {
			if parts[1] != "=>" {
				return nil, fmt.Errorf("expected path replacement, received: %s", l)
			}
			replacement = false
		}
		switch parts[1] {
		case "path":
			if len(parts) != 3 {
				return nil, fmt.Errorf("invalid path line: %s", l)
			}
			current.PackagePath = parts[2]
		case "mod":
			if len(parts) != 5 {
				return nil, fmt.Errorf("invalid mod line: %s", l)
			}
			current.ModulePath = parts[2]
		case "dep", "=>":
			switch len(parts) {
			case 5:
				current.ModuleRefs = append(current.ModuleRefs, model.ModuleReference{
					Path:    parts[2],
					Version: parts[3],
				})
			case 4:
				replacement = true
			default:
				return nil, fmt.Errorf("invalid dep line: %s", l)
			}
		case "build":
			// introduced in Go 1.18 - not captured as we aren't using it for anything
		case "":
			// blank (tab prefixed) lines appear after lines relating to replace directives in Go 1.18 compiled binaries
		default:
			return nil, fmt.Errorf("unrecognised line: %s", l)
		}
	}
	if current.Path != "" {
		results = append(results, current)
	}
	return results, nil
}
