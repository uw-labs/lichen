package model

import "fmt"

type Binary struct {
	Path string
	Refs []Reference
}

type Module struct {
	Reference
	Dir      string
	Licenses []License
}

type Reference struct {
	Path    string
	Version string
}

func (r Reference) String() string {
	return fmt.Sprintf("%s@%s", r.Path, r.Version)
}

type License struct {
	Name       string
	Confidence float64
}
