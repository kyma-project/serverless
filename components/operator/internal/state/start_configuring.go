package state

import (
	"context"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnStartConfiguring(_ context.Context, _ *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	s.setState(v1alpha1.StateProcessing)
	s.instance.Status.DockerRegistry = ""
	return nextState(sFnOptionalDependencies)
}
