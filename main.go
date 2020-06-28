package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/golly/internal/scan"
	"gopkg.in/yaml.v2"
)

const tmpl = `{{range .}}
{{- .Module.Path}}: {{range $i, $_ := .Module.Licenses}}{{if $i}}, {{end}}{{.Name}}{{end}} ({{.Explain}})
{{end}}`

func main() {
	a := &cli.App{
		Name: "golly",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "path to config file",
			},
			&cli.StringFlag{
				Name:  "template",
				Usage: "template for writing out each module and resolved licenses",
				Value: tmpl,
			},
			&cli.StringFlag{
				Name:  "json",
				Usage: "write JSON results to the supplied file",
			},
		},
		Action: run,
	}

	if err := a.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	output, err := template.New("output").Parse(c.String("template"))
	if err != nil {
		return err
	}

	conf, err := parseConfig(c.String("config"))
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	results, err := scan.Run(c.Context, conf, c.Args().Slice()...)
	if err != nil {
		return fmt.Errorf("failed to evaluate licenses: %w", err)
	}

	if jsonPath := c.String("json"); jsonPath != "" {
		if err := writeJSON(jsonPath, results); err != nil {
			return fmt.Errorf("failed to write json: %w", err)
		}
	}

	if err := output.Execute(os.Stdout, results); err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}

	var rErr error
	for _, res := range results {
		if !res.Allowed() {
			rErr = multierror.Append(rErr, fmt.Errorf("%s not allowed", res.Module.Path))
		}
	}
	return rErr
}

func parseConfig(path string) (scan.Config, error) {
	var conf scan.Config
	if path != "" {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return scan.Config{}, fmt.Errorf("failed to read file %q: %w", path, err)
		}
		if err := yaml.Unmarshal(b, &conf); err != nil {
			return scan.Config{}, fmt.Errorf("failed to parse yaml: %w", err)
		}
	}
	return conf, nil
}

func writeJSON(path string, results []scan.Result) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file for json output: %w", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(results); err != nil {
		return fmt.Errorf("json encode failed: %w", err)
	}
	return nil
}
