package state

import (
	"context"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/recorder"
)

func sFnConfigureNetworkPolicies(ctx context.Context, r *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
	updateNetworkPoliciesStatus(r.k8s, &s.instance)
	return nextState(sFnApplyResources)
}

func updateNetworkPoliciesStatus(eventRecorder recorder.EventRecorder, instance *v1alpha1.Serverless) {
	fields := fieldsToUpdate{
		{"True", &instance.Status.NetworkPoliciesEnabled, "NetworkPolicies enabled", ""},
	}
	updateStatusFields(eventRecorder, instance, fields)
}
