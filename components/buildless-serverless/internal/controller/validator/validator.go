package validator

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/config"
	"github.com/kyma-project/serverless/components/common/fips"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/api/validation"
	v1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type validator struct {
	instance *serverlessv1alpha2.Function
	fnConfig config.FunctionConfig
}

func New(instance *serverlessv1alpha2.Function, fnConfig config.FunctionConfig) *validator {
	return &validator{
		instance: instance,
		fnConfig: fnConfig,
	}
}

func (v *validator) Validate() []string {
	fns := []func() []string{
		v.validateEnvs,
		v.validateInlineDeps,
		v.validateRuntime,
		v.validateSecretMounts,
		v.validateFunctionLabels,
		v.validateFunctionAnnotations,
		v.validateGitRepoURL,
		v.validateFips,
		v.validateFunctionResources,
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

func (v *validator) validateFips() []string {
	var result []string
	if !fips.IsFIPS140Only() {
		return result
	}
	if err := validateSshGitIsForbiddenInFipsMode(v.instance.Spec.Source.GitRepository); err != nil {
		result = append(result, err.Error())
	}
	return result
}

func (v *validator) validateFunctionResources() []string {
	rc := v.instance.Spec.ResourceConfiguration
	minCPU := v.fnConfig.ResourceConfig.Function.Resources.MinRequestCPU.Quantity
	minMemory := v.fnConfig.ResourceConfig.Function.Resources.MinRequestMemory.Quantity
	if rc != nil && rc.Function != nil && rc.Function.Resources != nil {
		vrLimits := validateLimits(*rc.Function.Resources, minMemory, minCPU, "Function")
		vrRequests := validateRequests(*rc.Function.Resources, minMemory, minCPU, "Function")
		vr := append(vrLimits, vrRequests...)
		return vr
	}
	return []string{}
}

func validateDependencies(runtime serverlessv1alpha2.Runtime, dependencies string) error {
	if runtime.IsRuntimeNodejs() {
		return validateNodeJSDependencies(dependencies)
	}
	if runtime.IsRuntimePython() {
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
	if runtime.IsRuntimeKnown() {
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

func validateSshGitIsForbiddenInFipsMode(gitRepo *serverlessv1alpha2.GitRepositorySource) error {
	if gitRepo == nil {
		return nil
	}
	if urlIsSSH(gitRepo.URL) {
		return errors.New("SSH source.gitRepository.URL is not allowed in FIPS mode")
	}
	return nil
}

func validateLimits(resources corev1.ResourceRequirements, minMemory, minCPU resource.Quantity, resourceType string) []string {
	limits := resources.Limits
	allErrs := []string{}

	if limits != nil {
		if limits.Cpu().Cmp(minCPU) == -1 {
			allErrs = append(allErrs, fmt.Sprintf("%s limits cpu(%s) should be higher than minimal value (%s)",
				resourceType, limits.Cpu().String(), minCPU.String()))
		}
		if limits.Memory().Cmp(minMemory) == -1 {
			allErrs = append(allErrs, fmt.Sprintf("%s limits memory(%s) should be higher than minimal value (%s)",
				resourceType, limits.Memory().String(), minMemory.String()))
		}
	}
	return allErrs
}

func validateRequests(resources corev1.ResourceRequirements, minMemory, minCPU resource.Quantity, resourceType string) []string {
	limits := resources.Limits
	requests := resources.Requests
	allErrs := []string{}

	if requests != nil {
		if requests.Cpu().Cmp(minCPU) == -1 {
			allErrs = append(allErrs, fmt.Sprintf("%s request cpu(%s) should be higher than minimal value (%s)",
				resourceType, requests.Cpu().String(), minCPU.String()))
		}
		if requests.Memory().Cmp(minMemory) == -1 {
			allErrs = append(allErrs, fmt.Sprintf("%s request memory(%s) should be higher than minimal value (%s)",
				resourceType, requests.Memory().String(), minMemory.String()))
		}
	}

	if limits == nil {
		return allErrs
	}

	if requests.Cpu().Cmp(*limits.Cpu()) == 1 {
		allErrs = append(allErrs, fmt.Sprintf("%s limits cpu(%s) should be higher than requests cpu(%s)",
			resourceType, limits.Cpu().String(), requests.Cpu().String()))
	}
	if requests.Memory().Cmp(*limits.Memory()) == 1 {
		allErrs = append(allErrs, fmt.Sprintf("%s limits memory(%s) should be higher than requests memory(%s)",
			resourceType, limits.Memory().String(), requests.Memory().String()))
	}

	return allErrs
}
