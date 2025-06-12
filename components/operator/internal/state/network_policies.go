package state

import (
	"context"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	record "k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func sFnConfigureNetworkPolicies(ctx context.Context, r *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
	configureNetworkPolicyFlag(s)
	updateNetworkPoliciesStatus(r.k8s, &s.instance)
	return nextState(sFnApplyResources)
}

func configureNetworkPolicyFlag(s *systemState) {
	s.flagsBuilder.
		WithEnableNetworkPolicies(s.instance.Spec.EnableNetworkPolicies)
}

func updateNetworkPoliciesStatus(eventRecorder record.EventRecorder, instance *v1alpha1.Serverless) {
	stringValue := "False"
	if instance.Spec.EnableNetworkPolicies {
		stringValue = "True"
	}
	fields := fieldsToUpdate{
		{stringValue, &instance.Status.NetworkPoliciesEnabled, "NetworkPolicies enabled", ""},
	}
	updateStatusFields(eventRecorder, instance, fields)
}
