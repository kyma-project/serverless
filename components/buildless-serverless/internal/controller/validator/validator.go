package validator

import (
	"errors"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
	"strings"
)

type Validator interface {
	Validate() []string
}

type validator struct {
	instance *serverlessv1alpha2.Function
}

func New(instance *serverlessv1alpha2.Function) Validator {
	return &validator{
		instance: instance,
	}
}

func (v *validator) Validate() []string {
	fns := []func() []string{
		v.validateEnvs,
		v.validateInlineDeps,
		//TODO: add more validation functions
	}

	results := []string{}
	for _, fn := range fns {
		result := fn()
		results = append(results, result...)
	}
	return results
}

func (v *validator) validateEnvs() []string {
	for _, env := range v.instance.Spec.Env {
		vr := utilvalidation.IsEnvVarName(env.Name)
		if len(vr) != 0 {
			return enrichErrors(vr, "spec.env", env.Name)
		}
	}
	return []string{}
}

func (v *validator) validateInlineDeps() []string {
	runtime := v.instance.Spec.Runtime
	inlineSource := v.instance.Spec.Source.Inline
	if inlineSource == nil {
		return []string{}
	}
	if err := validateDependencies(runtime, inlineSource.Dependencies); err != nil {
		return []string{
			fmt.Sprintf("invalid source.inline.dependencies value: %s", err.Error()),
		}
	}
	return []string{}
}

func validateDependencies(runtime serverlessv1alpha2.Runtime, dependencies string) error {
	switch runtime {
	case serverlessv1alpha2.NodeJs20:
		return validateNodeJSDependencies(dependencies)
	case serverlessv1alpha2.Python312:
		return nil
	}
	return fmt.Errorf("cannot find runtime: %s", runtime)
}

func validateNodeJSDependencies(dependencies string) error {
	if deps := strings.TrimSpace(dependencies); deps != "" && (deps[0] != '{' || deps[len(deps)-1] != '}') {
		return errors.New("deps should start with '{' and end with '}'")
	}
	return nil
}

func enrichErrors(errs []string, path string, value string) []string {
	enrichedErrs := []string{}
	for _, err := range errs {
		enrichedErrs = append(enrichedErrs, fmt.Sprintf("%s: %s. Err: %s", path, value, err))
	}
	return enrichedErrs
}
