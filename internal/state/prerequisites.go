package state

import (
	"context"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/dependencies"
	ctrl "sigs.k8s.io/controller-runtime"
)

// check necessery dependencies before installation
func sFnPrerequisites() (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		// check hard serverless dependencies before installation
		withIstio := s.instance.Spec.DockerRegistry.IsInternalEnabled()
		err := dependencies.CheckPrerequisites(ctx, r.client, withIstio)
		if err != nil {
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		// when we know that cluster configuration met serverless requirements we can go to installation state
		return sFnApplyResources()
	}, nil, nil
}
