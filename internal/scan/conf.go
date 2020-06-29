package scan

type Config struct {
	Threshold  *float64    `yaml:"threshold"`
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
