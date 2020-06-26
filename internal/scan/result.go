package scan

import (
	"fmt"

	"github.com/utilitywarehouse/golly/internal/model"
)

type Result struct {
	Module       model.Module
	Decision     Decision
	NotPermitted []string `json:",omitempty"`
	Binaries     []string
}

func (r Result) Allowed() bool {
	return r.Decision == DecisionAllowed
}

func (r Result) Explain() string {
	switch r.Decision {
	case DecisionAllowed:
		return "allowed"
	case DecisionNotAllowedUnresolvableLicense:
		return "not allowed - unresolvable license"
	case DecisionNotAllowedLicenseNotPermitted:
		return fmt.Sprintf("not allowed - non-permitted licenses: %v", r.NotPermitted)
	default:
		panic("unrecognised decision")
	}
}

type Decision int

const (
	DecisionAllowed Decision = 1 + iota
	DecisionNotAllowedUnresolvableLicense
	DecisionNotAllowedLicenseNotPermitted
)

func (d Decision) MarshalText() ([]byte, error) {
	switch d {
	case DecisionAllowed:
		return []byte("allowed"), nil
	case DecisionNotAllowedUnresolvableLicense:
		return []byte("unresolvable-license"), nil
	case DecisionNotAllowedLicenseNotPermitted:
		return []byte("licenses-not-allowed"), nil
	default:
		panic("unrecognised decision")
	}
}
