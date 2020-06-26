package model

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

type License struct {
	Name       string
	Confidence float64
}
