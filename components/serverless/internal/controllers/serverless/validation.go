package serverless

import (
	"context"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type validationResult struct {
	valid bool
	msg   string
}

var _ stateFn = stateFnValidateFunction

func stateFnValidateFunction(_ context.Context, r *reconciler, s *systemState) (stateFn, error) {
	rc := s.instance.Spec.ResourceConfiguration
	fnResourceCfg := r.cfg.fn.ResourceConfig.Function.Resources
	vr := validateFunctionResources(rc, fnResourceCfg.MinRequestedCPU.Quantity, fnResourceCfg.MinRequestedMemory.Quantity)
	if !vr.valid {
		cond := createValidationFailedCondition(vr.msg)
		r.result.Requeue = false
		return buildStatusUpdateStateFnWithCondition(cond), nil
	}

	buildResourceCfg := r.cfg.fn.ResourceConfig.BuildJob.Resources
	vr = validateBuildResources(rc, buildResourceCfg.MinRequestedCPU.Quantity, buildResourceCfg.MinRequestedCPU.Quantity)
	if !vr.valid {
		cond := createValidationFailedCondition(vr.msg)
		r.result.Requeue = false
		return buildStatusUpdateStateFnWithCondition(cond), nil
	}
	return stateFnInitialize, nil
}

func validateFunctionResources(rc *serverlessv1alpha2.ResourceConfiguration, minCPU resource.Quantity, minMemory resource.Quantity) validationResult {
	if rc != nil && rc.Function != nil && rc.Function.Resources != nil {
		vr := validateResources(rc.Function.Resources, minCPU, minMemory)
		if !vr.valid {
			vr.msg = fmt.Sprintf("Function resources is not valid. %s", vr.msg)
		}
		return vr
	}
	return validationResult{valid: true}
}

func validateBuildResources(rc *serverlessv1alpha2.ResourceConfiguration, minCPU resource.Quantity, minMemory resource.Quantity) validationResult {
	if rc != nil && rc.Build != nil && rc.Build.Resources != nil {
		vr := validateResources(rc.Build.Resources, minCPU, minMemory)
		if !vr.valid {
			vr.msg = fmt.Sprintf("Build resources is not valid. %s", vr.msg)
		}
		return vr
	}
	return validationResult{valid: true}
}

func validateResources(r *corev1.ResourceRequirements, minCPU resource.Quantity, minMemory resource.Quantity) validationResult {
	if !r.Requests.Cpu().IsZero() && isMore(&minCPU, *r.Requests.Cpu()) {
		return validationResult{
			valid: false,
			msg:   fmt.Sprintf("Request cpu cannot be less than minimum cpu: %s", minCPU.String())}
	}

	if !r.Requests.Memory().IsZero() && isMore(&minMemory, *r.Requests.Memory()) {
		return validationResult{
			valid: false,
			msg:   fmt.Sprintf("Request memory cannot be less than minimum memory: %s", minMemory.String())}
	}

	if !r.Limits.Cpu().IsZero() && isMore(&minCPU, *r.Limits.Cpu()) {
		return validationResult{
			valid: false,
			msg:   fmt.Sprintf("Limits cpu cannot be less than minimum cpu: %s", minCPU.String())}
	}

	if !r.Limits.Memory().IsZero() && isMore(&minMemory, *r.Limits.Memory()) {
		return validationResult{
			valid: false,
			msg:   fmt.Sprintf("Limits memory cannot be less than minimum memory: %s", minMemory.String())}
	}

	if !r.Requests.Cpu().IsZero() && isMore(r.Requests.Cpu(), *r.Limits.Cpu()) {
		return validationResult{
			valid: false,
			msg:   "Request cpu cannot be bigger than limits cpu"}
	}

	if !r.Requests.Memory().IsZero() && isMore(r.Requests.Memory(), *r.Limits.Memory()) {
		return validationResult{
			valid: false,
			msg:   "Request memory cannot be bigger than limits memory"}
	}
	return validationResult{valid: true}
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
