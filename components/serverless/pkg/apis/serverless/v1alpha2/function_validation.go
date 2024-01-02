package v1alpha2

import (
	"fmt"
	"net/url"
	"regexp"
)

const ValidationConfigKey = "validation-config"

type MinFunctionResourcesValues struct {
	MinRequestCPU    string
	MinRequestMemory string
}

type MinBuildJobResourcesValues struct {
	MinRequestCPU    string
	MinRequestMemory string
}

type MinFunctionValues struct {
	Resources MinFunctionResourcesValues
}

type MinBuildJobValues struct {
	Resources MinBuildJobResourcesValues
}

type ValidationConfig struct {
	ReservedEnvs []string
	Function     MinFunctionValues
	BuildJob     MinBuildJobValues
}

type validationFunction func(*ValidationConfig) error

func (fn *Function) Validate(vc *ValidationConfig) error {
	if fn.TypeOf(FunctionTypeGit) {
		validations := fn.Spec.gitAuthValidations()
		return runValidations(vc, validations...)
	}
	return nil
}

func runValidations(vc *ValidationConfig, vFuns ...validationFunction) error {
	allErrs := []string{}
	for _, vFun := range vFuns {
		if err := vFun(vc); err != nil {
			allErrs = append(allErrs, err.Error())
		}
	}
	return returnAllErrs("", allErrs)
}

func (spec *FunctionSpec) validateGitRepoURL(_ *ValidationConfig) error {
	if urlIsSSH(spec.Source.GitRepository.URL) {
		return nil
	} else if _, err := url.ParseRequestURI(spec.Source.GitRepository.URL); err != nil {
		return fmt.Errorf("invalid source.gitRepository.URL value: %v", err)
	}
	return nil
}

func (spec *FunctionSpec) gitAuthValidations() []validationFunction {
	if spec.Source.GitRepository.Auth == nil {
		return []validationFunction{}
	}
	return []validationFunction{
		spec.validateGitRepoURL,
	}
}

func urlIsSSH(repoURL string) bool {
	exp, err := regexp.Compile(`((git|ssh?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(/)?`)
	if err != nil {
		panic(err)
	}

	return exp.MatchString(repoURL)
}
