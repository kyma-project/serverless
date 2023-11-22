package serverless

import (
	"context"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

var _ stateFn = stateFnValidateFunction

func stateFnValidateFunction(_ context.Context, r *reconciler, s *systemState) (stateFn, error) {
	rc := s.instance.Spec.ResourceConfiguration
	fnResourceCfg := r.cfg.fn.ResourceConfig.Function.Resources
	vrFunction := validateFunctionResources(rc, fnResourceCfg.MinRequestedCPU.Quantity, fnResourceCfg.MinRequestedMemory.Quantity)
	buildResourceCfg := r.cfg.fn.ResourceConfig.BuildJob.Resources
	vrBuild := validateBuildResources(rc, buildResourceCfg.MinRequestedCPU.Quantity, buildResourceCfg.MinRequestedCPU.Quantity)

	vr := append(vrFunction, vrBuild...)
	if len(vr) != 0 {
		msg := strings.Join(vr, ".")
		cond := createValidationFailedCondition(msg)
		r.result.Requeue = false
		return buildStatusUpdateStateFnWithCondition(cond), nil
	}
	return stateFnInitialize, nil
}

func validateFunctionResources(rc *serverlessv1alpha2.ResourceConfiguration, minCPU resource.Quantity, minMemory resource.Quantity) []string {
	if rc != nil && rc.Function != nil && rc.Function.Resources != nil {
		vrLimits := validateLimits(*rc.Function.Resources, minMemory, minCPU, "Function")
		vrRequests := validateRequests(*rc.Function.Resources, minMemory, minCPU, "Function")
		vr := append(vrLimits, vrRequests...)
		return vr
	}
	return []string{}
}

func validateBuildResources(rc *serverlessv1alpha2.ResourceConfiguration, minCPU resource.Quantity, minMemory resource.Quantity) []string {
	if rc != nil && rc.Build != nil && rc.Build.Resources != nil {
		vrLimits := validateLimits(*rc.Build.Resources, minMemory, minCPU, "Function")
		vrRequests := validateRequests(*rc.Build.Resources, minMemory, minCPU, "Function")
		vr := append(vrLimits, vrRequests...)
		return vr
	}
	return []string{}
}

func validateRequests(resources corev1.ResourceRequirements, minMemory, minCPU resource.Quantity, resourceType string) []string {
	limits := resources.Limits
	requests := resources.Requests
	allErrs := []string{}

	if requests.Cpu().Cmp(minCPU) == -1 {
		allErrs = append(allErrs, fmt.Sprintf("%s request cpu(%s) should be higher than minimal value (%s)",
			resourceType, requests.Cpu().String(), minCPU.String()))
	}
	if requests.Memory().Cmp(minMemory) == -1 {
		allErrs = append(allErrs, fmt.Sprintf("%s request memory(%s) should be higher than minimal value (%s)",
			resourceType, requests.Memory().String(), minMemory.String()))
	}

	if limits == nil {
		return allErrs
	}

	if requests.Cpu().Cmp(*limits.Cpu()) == 1 {
		allErrs = append(allErrs, fmt.Sprintf("%s limits cpu(%s) should be higher than Requests cpu(%s)",
			resourceType, limits.Cpu().String(), requests.Cpu().String()))
	}
	if requests.Memory().Cmp(*limits.Memory()) == 1 {
		allErrs = append(allErrs, fmt.Sprintf("%s limits memory(%s) should be higher than Requests.memory(%s)",
			resourceType, limits.Memory().String(), requests.Memory().String()))
	}

	return allErrs
}

func validateLimits(resources corev1.ResourceRequirements, minMemory, minCPU resource.Quantity, resourceType string) []string {
	limits := resources.Limits
	allErrs := []string{}

	if limits != nil {
		if limits.Cpu().Cmp(minCPU) == -1 {
			allErrs = append(allErrs, fmt.Sprintf("%s limits.cpu(%s) should be higher than minimal value (%s)",
				resourceType, limits.Cpu().String(), minCPU.String()))
		}
		if limits.Memory().Cmp(minMemory) == -1 {
			allErrs = append(allErrs, fmt.Sprintf("%s limits.memory(%s) should be higher than minimal value (%s)",
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

func isMore(q1 *resource.Quantity, q2 resource.Quantity) bool {
	return q1.Cmp(q2) != -1
}
