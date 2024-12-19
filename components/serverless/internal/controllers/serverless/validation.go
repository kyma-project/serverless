package serverless

import (
	"context"
	"fmt"
	"strings"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ stateFn = stateFnValidateFunction

type validationFn func() []string

func stateFnValidateFunction(_ context.Context, r *reconciler, s *systemState) (stateFn, error) {
	rc := s.instance.Spec.ResourceConfiguration
	fnResourceCfg := r.cfg.fn.ResourceConfig.Function.Resources
	validateFunctionResources := validateFunctionResourcesFn(rc, fnResourceCfg.MinRequestedCPU.Quantity, fnResourceCfg.MinRequestedMemory.Quantity)
	buildResourceCfg := r.cfg.fn.ResourceConfig.BuildJob.Resources
	validateBuildResources := validateBuildResourcesFn(rc, buildResourceCfg.MinRequestedCPU.Quantity, buildResourceCfg.MinRequestedMemory.Quantity)

	spec := s.instance.Spec
	validationFns := []validationFn{
		validateRuntime(spec.Runtime),
		validateFunctionResources,
		validateBuildResources,
		validateEnvs(spec.Env, "spec.env"),
		validateSecretMounts(spec.SecretMounts),
		validateInlineDeps(spec.Runtime, spec.Source.Inline),
		validateFunctionLabels(spec.Labels, "spec.labels"),
		validateFunctionAnnotations(spec.Annotations, "spec.annotations"),
		validateGitRepoURL(spec.Source),
	}
	validationResults := []string{}
	for _, validationFn := range validationFns {
		result := validationFn()
		validationResults = append(validationResults, result...)
	}

	if len(validationResults) != 0 {
		msg := strings.Join(validationResults, ". ")
		cond := createValidationFailedCondition(msg)
		r.result = ctrl.Result{Requeue: false}
		return buildStatusUpdateStateFnWithCondition(cond), nil
	}
	return stateFnInitialize, nil
}

func validateGitRepoURL(source serverlessv1alpha2.Source) validationFn {
	return func() []string {
		var result []string
		if source.GitRepository == nil {
			return result
		}
		if err := serverlessv1alpha2.ValidateGitRepoURL(source.GitRepository); err != nil {
			result = append(result, err.Error())
		}
		return result
	}
}

func validateFunctionResourcesFn(rc *serverlessv1alpha2.ResourceConfiguration, minCPU resource.Quantity, minMemory resource.Quantity) validationFn {
	return func() []string {
		if rc != nil && rc.Function != nil && rc.Function.Resources != nil {
			vrLimits := validateLimits(*rc.Function.Resources, minMemory, minCPU, "Function")
			vrRequests := validateRequests(*rc.Function.Resources, minMemory, minCPU, "Function")
			vr := append(vrLimits, vrRequests...)
			return vr
		}
		return []string{}
	}
}

func validateBuildResourcesFn(rc *serverlessv1alpha2.ResourceConfiguration, minCPU resource.Quantity, minMemory resource.Quantity) validationFn {
	return func() []string {
		if rc != nil && rc.Build != nil && rc.Build.Resources != nil {
			vrLimits := validateLimits(*rc.Build.Resources, minMemory, minCPU, "Build")
			vrRequests := validateRequests(*rc.Build.Resources, minMemory, minCPU, "Build")
			vr := append(vrLimits, vrRequests...)
			return vr
		}
		return []string{}
	}
}

func validateEnvs(envs []corev1.EnvVar, path string) validationFn {
	return func() []string {
		for _, env := range envs {
			vr := utilvalidation.IsEnvVarName(env.Name)
			if len(vr) != 0 {
				return enrichErrors(vr, path, env.Name)
			}
		}
		return []string{}
	}
}

func validateSecretMounts(secretMounts []serverlessv1alpha2.SecretMount) validationFn {
	return func() []string {
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
}

func secretNamesAreUnique(secretMounts []serverlessv1alpha2.SecretMount) bool {
	uniqueSecretNames := make(map[string]bool)
	for _, secretMount := range secretMounts {
		uniqueSecretNames[secretMount.SecretName] = true
	}
	return len(uniqueSecretNames) == len(secretMounts)
}

func validateInlineDeps(runtime serverlessv1alpha2.Runtime, inlineSource *serverlessv1alpha2.InlineSource) validationFn {
	return func() []string {
		if inlineSource == nil {
			return []string{}
		}
		if err := serverlessv1alpha2.ValidateDependencies(runtime, inlineSource.Dependencies); err != nil {
			return []string{
				fmt.Sprintf("invalid source.inline.dependencies value: %s", err.Error()),
			}
		}
		return []string{}
	}
}

func validateRuntime(runtime serverlessv1alpha2.Runtime) validationFn {
	return func() []string {
		if err := serverlessv1alpha2.ValidateRuntime(runtime); err != nil {
			return []string{
				fmt.Sprintf("invalid runtime value: %s", err.Error()),
			}
		}
		return []string{}
	}
}

func validateFunctionLabels(labels map[string]string, path string) validationFn {
	return func() []string {
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
}

func validateFunctionAnnotations(annotations map[string]string, path string) validationFn {
	return func() []string {
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
}

func enrichErrors(errs []string, path string, value string) []string {
	enrichedErrs := []string{}
	for _, err := range errs {
		enrichedErrs = append(enrichedErrs, fmt.Sprintf("%s: %s. Err: %s", path, value, err))
	}
	return enrichedErrs
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

func createValidationFailedCondition(msg string) serverlessv1alpha2.Condition {
	return serverlessv1alpha2.Condition{
		Type:               serverlessv1alpha2.ConditionConfigurationReady,
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha2.ConditionReasonFunctionSpec,
		Message:            msg,
	}
}
