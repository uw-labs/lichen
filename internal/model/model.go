package model

import "fmt"

// BuildInfo encapsulates build info embedded into a Go compile binary
type BuildInfo struct {
	Path        string            // OS level absolute path to the binary this build info relates to
	PackagePath string            // package path indicated by the build info, e.g. github.com/foo/bar/cmd/baz
	ModulePath  string            // module path indicated by the build info, e.g. github.com/foo/bar
	ModuleRefs  []ModuleReference // all modules that feature in the build info output
}

// Module carries details of a Go module
type Module struct {
	ModuleReference           // reference (path & version)
	Dir             string    // OS level absolute path to where the cached copy of the module is located
	Licenses        []License // resolved licenses
}

// ModuleReference is a reference to a particular version of a named module
type ModuleReference struct {
	Path    string // module path, e.g. github.com/foo/bar
	Version string // module version (can take a variety of forms)
}

// String returns a typical string representation of a module reference (path@version)
func (r ModuleReference) String() string {
	return fmt.Sprintf("%s@%s", r.Path, r.Version)
}

// License carries license classification details
type License struct {
	Path       string  // OS level absolute path to the license file
	Name       string  // SPDX name of the license
	Confidence float64 // confidence from license classification
}
