package scan

import (
	"context"
	"sort"

	"github.com/uw-labs/lichen/internal/license"
	"github.com/uw-labs/lichen/internal/model"
	"github.com/uw-labs/lichen/internal/module"
)

const defaultThreshold = 0.80

func Run(ctx context.Context, conf Config, binPaths ...string) (Summary, error) {
	// extract modules details from each supplied binary
	binaries, err := module.Extract(ctx, binPaths...)
	if err != nil {
		return Summary{}, err
	}

	// fetch each module - this returns pertinent details, including the OS path to the module
	modules, err := module.Fetch(ctx, uniqueModuleRefs(binaries))
	if err != nil {
		return Summary{}, err
	}

	// resolve licenses based on a minimum threshold
	threshold := defaultThreshold
	if conf.Threshold != nil {
		threshold = *conf.Threshold
	}
	modules, err = license.Resolve(modules, threshold)
	if err != nil {
		return Summary{}, err
	}

	// apply any overrides, if configured
	if len(conf.Overrides) > 0 {
		modules = applyOverrides(modules, conf.Overrides)
	}

	// evaluate the modules and sort by path
	results := evaluate(conf, binaries, modules)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Module.Path < results[j].Module.Path
	})

	return Summary{
		Binaries: binaries,
		Modules:  results,
	}, nil
}

// uniqueModuleRefs returns all unique modules referenced by the supplied binaries
func uniqueModuleRefs(infos []model.BuildInfo) []model.ModuleReference {
	unique := make(map[model.ModuleReference]struct{})
	for _, res := range infos {
		for _, r := range res.ModuleRefs {
			unique[r] = struct{}{}
		}
	}

	refs := make([]model.ModuleReference, 0, len(unique))
	for r := range unique {
		refs = append(refs, r)
	}

	return refs
}

// applyOverrides replaces license information
func applyOverrides(modules []model.Module, overrides []Override) []model.Module {
	replacements := make(map[string][]string, len(overrides))
	for _, o := range overrides {
		replacements[o.Path] = o.Licenses
	}

	for i, mod := range modules {
		if repl, found := replacements[mod.ModuleReference.Path]; found {
			mod.Licenses = make([]model.License, 0, len(repl))
			for _, lic := range repl {
				mod.Licenses = append(mod.Licenses, model.License{
					Name:       lic,
					Confidence: 1,
				})
			}
			modules[i] = mod
		}
	}

	return modules
}

// evaluate inspects each module, checking that (a) license details could be determined, and (b) licenses
// are permitted by the supplied configuration.
func evaluate(conf Config, binaries []model.BuildInfo, modules []model.Module) []EvaluatedModule {
	// build a map each module to binaries that reference them
	binRefs := make(map[model.ModuleReference][]string, len(modules))
	for _, bin := range binaries {
		for _, ref := range bin.ModuleRefs {
			binRefs[ref] = append(binRefs[ref], bin.Path)
		}
	}

	// build a map of permitted licenses
	permitted := make(map[string]bool, len(conf.Allow))
	for _, lic := range conf.Allow {
		permitted[lic] = true
	}

	// build a map of exceptions, based on path and license
	type pathLicense struct {
		path    string
		license string
	}
	ignore := make(map[pathLicense]bool, len(conf.Exceptions))
	for _, exception := range conf.Exceptions {
		for _, lic := range exception.Licenses {
			pl := pathLicense{
				path:    exception.Path,
				license: lic,
			}
			ignore[pl] = true
		}
	}

	// check each module
	results := make([]EvaluatedModule, 0, len(modules))
	for _, mod := range modules {
		res := EvaluatedModule{
			Module:   mod,
			UsedBy:   binRefs[mod.ModuleReference],
			Decision: DecisionAllowed,
		}
		if len(mod.Licenses) == 0 {
			res.Decision = DecisionNotAllowedUnresolvableLicense
		}
		for _, lic := range mod.Licenses {
			pl := pathLicense{
				path:    mod.Path,
				license: lic.Name,
			}
			if len(permitted) > 0 && !permitted[lic.Name] && !ignore[pl] {
				res.Decision = DecisionNotAllowedLicenseNotPermitted
				res.NotPermitted = append(res.NotPermitted, lic.Name)
			}
		}
		results = append(results, res)
	}
	return results
}
