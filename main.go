package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/go-multierror"
	"github.com/muesli/termenv"
	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/lichen/internal/scan"
	"gopkg.in/yaml.v2"
)

const tmpl = `{{range .}}
{{- .Module.Path}}: {{range $i, $_ := .Module.Licenses}}{{if $i}}, {{end}}{{.Name}}{{end}} 
{{- if .Allowed}} ({{ Color "#00ff00" .ExplainDecision}}){{else}} ({{ Color "#ff0000" .ExplainDecision}}){{end}}
{{end}}`

func main() {
	a := &cli.App{
		Name:  "lichen",
		Usage: "evaluate module dependencies from go compiled binaries",
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
	if c.NArg() == 0 {
		_ = cli.ShowAppHelp(c)
		return errors.New("path to at least one binary must be supplied")
	}

	f := termenv.TemplateFuncs(termenv.ColorProfile())
	output, err := template.New("output").Funcs(f).Parse(c.String("template"))
	if err != nil {
		return err
	}

	conf, err := parseConfig(c.String("config"))
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	paths, err := absolutePaths(c.Args().Slice())
	if err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}

	results, err := scan.Run(c.Context, conf, paths...)
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
			rErr = multierror.Append(rErr, fmt.Errorf("%s: %s", res.Module.Path, res.ExplainDecision()))
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

func absolutePaths(paths []string) ([]string, error) {
	mapped := make([]string, len(paths))
	for i, path := range paths {
		abs, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}
		mapped[i] = abs
	}
	return mapped, nil
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
