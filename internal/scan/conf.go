package scan

type Config struct {
	Threshold  *float64   `yaml:"threshold"`
	Allow      []string   `yaml:"allow"`
	Exceptions Exceptions `yaml:"exceptions"`
	Overrides  []Override `yaml:"override"`
}

type Exceptions struct {
	LicenseNotPermitted []LicenseNotPermitted `yaml:"licenseNotPermitted"`
	UnresolvableLicense []UnresolvableLicense `yaml:"unresolvableLicense"`
}

type LicenseNotPermitted struct {
	Path     string   `yaml:"path"`
	Version  string   `yaml:"version"`
	Licenses []string `yaml:"licenses"`
}

type UnresolvableLicense struct {
	Path    string `yaml:"path"`
	Version string `yaml:"version"`
}

type Override struct {
	Path     string   `yaml:"path"`
	Licenses []string `yaml:"licenses"`
}
