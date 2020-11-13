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
	type replacement struct {
		version  string
		licenses []string
	}
	replacements := make(map[string]replacement, len(overrides))
	for _, o := range overrides {
		replacements[o.Path] = replacement{
			version:  o.Version,
			licenses: o.Licenses,
		}
	}

	for i, mod := range modules {
		if repl, found := replacements[mod.ModuleReference.Path]; found {
			// if an explicit version is configured, only apply the override if the module version matches
			if repl.version != "" && repl.version != mod.Version {
				continue
			}
			mod.Licenses = make([]model.License, 0, len(repl.licenses))
			for _, lic := range repl.licenses {
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

	// check each module
	results := make([]EvaluatedModule, 0, len(modules))
	for _, mod := range modules {
		res := EvaluatedModule{
			Module:   mod,
			UsedBy:   binRefs[mod.ModuleReference],
			Decision: DecisionAllowed,
		}
		if len(mod.Licenses) == 0 && !ignoreUnresolvable(conf, mod) {
			res.Decision = DecisionNotAllowedUnresolvableLicense
		}
		for _, lic := range mod.Licenses {
			if len(permitted) > 0 && !permitted[lic.Name] && !ignoreNotPermitted(conf, mod, lic) {
				res.Decision = DecisionNotAllowedLicenseNotPermitted
				res.NotPermitted = append(res.NotPermitted, lic.Name)
			}
		}
		results = append(results, res)
	}
	return results
}

func ignoreUnresolvable(conf Config, mod model.Module) bool {
	for _, exception := range conf.Exceptions.UnresolvableLicense {
		if exception.Path == mod.Path && (exception.Version == "" || exception.Version == mod.Version) {
			return true
		}
	}
	return false
}

func ignoreNotPermitted(conf Config, mod model.Module, lic model.License) bool {
	for _, exception := range conf.Exceptions.LicenseNotPermitted {
		if exception.Path == mod.Path && (exception.Version == "" || exception.Version == mod.Version) {
			if len(exception.Licenses) == 0 {
				return true
			}
			for _, exceptionLicense := range exception.Licenses {
				if exceptionLicense == lic.Name {
					return true
				}
			}
		}
	}
	return false
}
