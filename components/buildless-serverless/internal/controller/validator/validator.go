package validator

import (
	"errors"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"k8s.io/apimachinery/pkg/api/validation"
	v1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

type validator struct {
	instance *serverlessv1alpha2.Function
}

func New(instance *serverlessv1alpha2.Function) *validator {
	return &validator{
		instance: instance,
	}
}

func (v *validator) Validate() []string {
	fns := []func() []string{
		v.validateEnvs,
		v.validateInlineDeps,
		v.validateRuntime,
		//TODO: add more validation functions
		v.validateSecretMounts,
		v.validateFunctionLabels,
		v.validateFunctionAnnotations,
		v.validateGitRepoURL,
	}

	r := []string{}
	for _, f := range fns {
		result := f()
		r = append(r, result...)
	}
	return r
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

func (v *validator) validateRuntime() []string {
	runtime := v.instance.Spec.Runtime

	if err := validateRuntime(runtime); err != nil {
		return []string{
			fmt.Sprintf("invalid runtime value: %s", err.Error()),
		}
	}
	return []string{}
}

func (v *validator) validateSecretMounts() []string {
	secretMounts := v.instance.Spec.SecretMounts
	var allErrs []string
	for _, secretMount := range secretMounts {
		allErrs = append(allErrs,
			utilvalidation.IsDNS1123Subdomain(secretMount.SecretName)...)
	}
	if !secretNamesAreUnique(secretMounts) {
		allErrs = append(allErrs, "secretNames should be unique")
	}
	if len(allErrs) == 0 {
		return []string{}
	}
	return []string{
		fmt.Sprintf("invalid spec.secretMounts: %s", allErrs),
	}
}

func (v *validator) validateFunctionLabels() []string {
	labels := v.instance.Spec.Labels
	path := "spec.labels"
	errs := field.ErrorList{}
	fieldPath := field.NewPath(path)
	errs = append(errs, v1validation.ValidateLabels(labels, fieldPath)...)
	if len(errs) == 0 {
		return []string{}
	}
	result := []string{}
	for _, err := range errs {
		if err != nil {
			result = append(result, err.Error())
		}
	}
	return result
}

func (v *validator) validateFunctionAnnotations() []string {
	annotations := v.instance.Spec.Annotations
	path := "spec.annotations"
	errs := field.ErrorList{}
	fieldPath := field.NewPath(path)
	errs = append(errs, validation.ValidateAnnotations(annotations, fieldPath)...)
	if len(errs) == 0 {
		return []string{}
	}
	result := []string{}
	for _, err := range errs {
		if err != nil {
			result = append(result, err.Error())
		}
	}
	return result
}

func (v *validator) validateGitRepoURL() []string {
	var result []string
	if v.instance.Spec.Source.GitRepository == nil {
		return result
	}
	if err := validateGitRepoURL(v.instance.Spec.Source.GitRepository); err != nil {
		result = append(result, err.Error())
	}
	return result
}

func validateDependencies(runtime serverlessv1alpha2.Runtime, dependencies string) error {
	switch runtime {
	case serverlessv1alpha2.NodeJs20, serverlessv1alpha2.NodeJs22:
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

func validateRuntime(runtime serverlessv1alpha2.Runtime) error {
	if len(runtime) == 0 {
		return nil
	}
	supportedruntimes := []serverlessv1alpha2.Runtime{serverlessv1alpha2.NodeJs20, serverlessv1alpha2.NodeJs22, serverlessv1alpha2.Python312}
	if slices.Contains(supportedruntimes, runtime) {
		return nil
	}
	return fmt.Errorf("cannot find runtime: %s", runtime)
}

func secretNamesAreUnique(secretMounts []serverlessv1alpha2.SecretMount) bool {
	uniqueSecretNames := make(map[string]bool)
	for _, secretMount := range secretMounts {
		uniqueSecretNames[secretMount.SecretName] = true
	}
	return len(uniqueSecretNames) == len(secretMounts)
}

func enrichErrors(errs []string, path string, value string) []string {
	enrichedErrs := []string{}
	for _, err := range errs {
		enrichedErrs = append(enrichedErrs, fmt.Sprintf("%s: %s. Err: %s", path, value, err))
	}
	return enrichedErrs
}

func urlIsSSH(repoURL string) bool {
	exp, err := regexp.Compile(`((git|ssh)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(/)?`)
	if err != nil {
		panic(err)
	}

	return exp.MatchString(repoURL)
}

func validateGitRepoURL(gitRepo *serverlessv1alpha2.GitRepositorySource) error {
	if urlIsSSH(gitRepo.URL) {
		return nil
	} else if _, err := url.ParseRequestURI(gitRepo.URL); err != nil {
		return fmt.Errorf("source.gitRepository.URL: %v", err)
	}
	return nil
}
