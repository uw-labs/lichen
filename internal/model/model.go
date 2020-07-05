package model

import "fmt"

type Binary struct {
	Path       string
	ModuleRefs []ModuleReference
}

type Module struct {
	ModuleReference
	Dir      string
	Licenses []License
}

type ModuleReference struct {
	Path    string
	Version string
}

func (r ModuleReference) String() string {
	return fmt.Sprintf("%s@%s", r.Path, r.Version)
}

type License struct {
	Name       string
	Confidence float64
}
