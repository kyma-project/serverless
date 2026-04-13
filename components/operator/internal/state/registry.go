package state

import (
	"context"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnRegistryConfiguration(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	s.setState(v1alpha1.StateProcessing)
	// setup status.dockerRegistry and set possible warnings
	err := configureRegistry(ctx, r, s)
	if err != nil {
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigurationErr,
			err,
		)
		return stopWithEventualError(err)
	}

	return nextState(sFnOptionalDependencies)
}

func configureRegistry(_ context.Context, _ *reconciler, s *systemState) error {
	s.instance.Status.DockerRegistry = ""
	return nil
}
