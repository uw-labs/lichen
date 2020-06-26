package scan

type Config struct {
	Allow      []string    `yaml:"allow"`
	Exceptions []Exception `yaml:"exceptions"`
	Overrides  []Override  `yaml:"override"`
}

type Exception struct {
	Path     string
	Licenses []string
}

type Override struct {
	Path     string
	Licenses []string
}
