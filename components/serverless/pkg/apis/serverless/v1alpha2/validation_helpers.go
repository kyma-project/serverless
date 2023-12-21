package v1alpha2

import (
	"fmt"
)

func returnAllErrs(msg string, allErrs []string) error {
	if len(allErrs) == 0 {
		return nil
	}

	if len(msg) > 0 {
		return fmt.Errorf("%s: %v", msg, allErrs)
	}

	return fmt.Errorf("%v", allErrs)
}
