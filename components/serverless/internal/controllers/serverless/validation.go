package serverless

import (
	"context"
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
	vr := validateFunctionResources(rc)
	if !vr.valid {
		cond := createValidationFailedCondition(vr.msg)
		r.result.Requeue = false
		return buildStatusUpdateStateFnWithCondition(cond), nil
	}

	vr = validateBuildResources(rc)
	if !vr.valid {
		cond := createValidationFailedCondition(vr.msg)
		r.result.Requeue = false
		return buildStatusUpdateStateFnWithCondition(cond), nil
	}
	return stateFnInitialize, nil
}

func validateFunctionResources(rc *serverlessv1alpha2.ResourceConfiguration) validationResult {
	if rc != nil && rc.Function != nil && rc.Function.Resources != nil {
		vr := validateResources(rc.Function.Resources)
		return vr
	}
	return validationResult{valid: true}
}

func validateBuildResources(rc *serverlessv1alpha2.ResourceConfiguration) validationResult {
	if rc != nil && rc.Build != nil && rc.Build.Resources != nil {
		vr := validateResources(rc.Build.Resources)
		return vr
	}
	return validationResult{valid: true}
}

func validateResources(r *corev1.ResourceRequirements) validationResult {
	if isMore(r.Requests.Cpu(), *r.Limits.Cpu()) {
		return validationResult{
			valid: false,
			msg:   "Request CPU cannot be bigger than Limits CPU"}
	}

	if isMore(r.Requests.Cpu(), *r.Limits.Cpu()) {
		return validationResult{
			valid: false,
			msg:   "Request memory cannot be bigger than Limits memory"}
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
