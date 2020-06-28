package scan

import (
	"context"

	"github.com/utilitywarehouse/golly/internal/dep"
	"github.com/utilitywarehouse/golly/internal/license"
	"github.com/utilitywarehouse/golly/internal/model"
)

func Run(ctx context.Context, conf Config, paths ...string) ([]Result, error) {
	binaries, err := dep.Extract(ctx, paths...)
	if err != nil {
		return nil, err
	}

	modules, err := dep.Fetch(ctx, uniqueModuleRefs(binaries))
	if err != nil {
		return nil, err
	}

	modules, err = license.Resolve(modules, .75)
	if err != nil {
		return nil, err
	}

	if len(conf.Overrides) > 0 {
		modules = applyOverrides(modules, conf.Overrides)
	}

	return evaluate(conf, binaries, modules), nil
}

func uniqueModuleRefs(infos []model.Binary) []model.Reference {
	unique := make(map[model.Reference]struct{})
	for _, res := range infos {
		for _, r := range res.Refs {
			unique[r] = struct{}{}
		}
	}

	refs := make([]model.Reference, 0, len(unique))
	for r := range unique {
		refs = append(refs, r)
	}

	return refs
}

func applyOverrides(modules []model.Module, overrides []Override) []model.Module {
	replacements := make(map[string][]string, len(overrides))
	for _, o := range overrides {
		replacements[o.Path] = o.Licenses
	}

	for i, mod := range modules {
		if repl, found := replacements[mod.Reference.Path]; found {
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

func evaluate(conf Config, binaries []model.Binary, modules []model.Module) []Result {
	binRefs := make(map[model.Reference][]string, len(modules))
	for _, bin := range binaries {
		for _, ref := range bin.Refs {
			binRefs[ref] = append(binRefs[ref], bin.Path)
		}
	}

	permitted := make(map[string]bool, len(conf.Allow))
	for _, lic := range conf.Allow {
		permitted[lic] = true
	}

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

	results := make([]Result, 0, len(modules))
	for _, mod := range modules {
		res := Result{
			Module:   mod,
			Binaries: binRefs[mod.Reference],
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
