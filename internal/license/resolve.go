package license

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/licenseclassifier"
	"github.com/uw-labs/lichen/internal/license/db"
	"github.com/uw-labs/lichen/internal/model"
)

// Resolve inspects each module and determines what it is licensed under. The returned slice contains each
// module enriched with license information.
func Resolve(modules []model.Module, threshold float64) ([]model.Module, error) {
	archiveFn := licenseclassifier.ArchiveFunc(func() ([]byte, error) {
		f, err := db.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open license databse: %w", err)
		}
		defer f.Close()
		return ioutil.ReadAll(f)
	})

	lc, err := licenseclassifier.New(threshold, archiveFn)
	if err != nil {
		return nil, err
	}

	for i, m := range modules {
		paths, err := locateLicenses(m.Dir)
		if err != nil {
			return nil, err
		}
		licenses, err := classify(lc, paths)
		if err != nil {
			return nil, err
		}
		m.Licenses = licenses
		modules[i] = m
	}

	return modules, nil
}

var fileRgx = regexp.MustCompile(`(?i)^li[cs]en[cs]e`)

// locateLicenses searches for license files
func locateLicenses(path string) (lp []string, err error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if !f.IsDir() && fileRgx.MatchString(f.Name()) && !strings.HasSuffix(f.Name(), ".go") {
			lp = append(lp, filepath.Join(path, f.Name()))
		}
	}
	return lp, nil
}

// classify inspects each license file and classifies it
func classify(lc *licenseclassifier.License, paths []string) ([]model.License, error) {
	licenses := make([]model.License, 0)
	for _, p := range paths {
		content, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, err
		}
		hits := make(map[string]float64)
		matches := lc.MultipleMatch(string(content), true)
		for _, match := range matches {
			if conf, found := hits[match.Name]; !found || match.Confidence > conf {
				hits[match.Name] = match.Confidence
			}
		}
		for name, confidence := range hits {
			licenses = append(licenses, model.License{
				Name:       name,
				Path:       p,
				Confidence: confidence,
			})
		}
	}
	return licenses, nil
}
