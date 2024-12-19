package v1alpha2

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

func ValidateDependencies(runtime Runtime, dependencies string) error {
	switch runtime {
	case NodeJs20, NodeJs22:
		return validateNodeJSDependencies(dependencies)
	case Python312:
		return nil
	}
	return nil
}

func validateNodeJSDependencies(dependencies string) error {
	if deps := strings.TrimSpace(dependencies); deps != "" && (deps[0] != '{' || deps[len(deps)-1] != '}') {
		return errors.New("deps should start with '{' and end with '}'")
	}
	return nil
}

func ValidateRuntime(runtime Runtime) error {
	if len(runtime) == 0 {
		return nil
	}
	supportedruntimes := []Runtime{NodeJs20, NodeJs22, Python312}
	if slices.Contains(supportedruntimes, runtime) {
		return nil
	}
	return fmt.Errorf("cannot find runtime: %s", runtime)
}
