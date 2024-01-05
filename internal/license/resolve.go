package license

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	classifier "github.com/google/licenseclassifier/v2"
	"github.com/google/licenseclassifier/v2/assets"
	"github.com/uw-labs/lichen/internal/model"
)

// Resolve inspects each module and determines what it is licensed under. The returned slice contains each
// module enriched with license information.
func Resolve(modules []model.Module, threshold float64) ([]model.Module, error) {
	lc, err := assets.DefaultClassifier()
	if err != nil {
		return nil, err
	}

	for i, m := range modules {
		if m.IsLocal() {
			// there is no guarantee we are being run in a location that makes local module references resolvable.. to
			// avoid incidental and non-obvious behaviour here, we simply don't touch such references - overrides must
			// be provided instead.
			continue
		}
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

var fileRgx = regexp.MustCompile(`(?i)^(li[cs]en[cs]e|copying)`)

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
func classify(lc *classifier.Classifier, paths []string) ([]model.License, error) {
	licenses := make([]model.License, 0)
	for _, p := range paths {
		content, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, err
		}
		hits := make(map[string]float64)
		matches := lc.Match(content)
		for _, match := range matches.Matches {
			if conf, found := hits[match.Name]; match.MatchType != "Copyright" && (!found || match.Confidence > conf) {
				hits[match.Name] = match.Confidence
			}
		}
		for name, confidence := range hits {
			licenses = append(licenses, model.License{
				Name:       name,
				Path:       p,
				Content:    string(content),
				Confidence: confidence,
			})
		}
	}
	return licenses, nil
}
