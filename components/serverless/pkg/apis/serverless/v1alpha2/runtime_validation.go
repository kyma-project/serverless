package v1alpha2

import (
	"errors"
	"fmt"
	"strings"
)

func ValidateDependencies(runtime Runtime, dependencies string) error {
	switch runtime {
	case NodeJs20, NodeJs22:
		return validateNodeJSDependencies(dependencies)
	case Python312:
		return nil
	}
	return fmt.Errorf("cannot find runtime: %s", runtime)
}

func ValidateRuntime(runtime Runtime) error {
	if len(runtime) > 0 && runtime != NodeJs20 && runtime != NodeJs22 && runtime != Python312 {
		return fmt.Errorf("cannot find runtime: %s", runtime)
	}
	return nil
}

func validateNodeJSDependencies(dependencies string) error {
	if deps := strings.TrimSpace(dependencies); deps != "" && (deps[0] != '{' || deps[len(deps)-1] != '}') {
		return errors.New("deps should start with '{' and end with '}'")
	}
	return nil
}
