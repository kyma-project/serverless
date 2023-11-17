package state

import (
	"context"

	"github.com/kyma-project/serverless-manager/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/components/operator/internal/gitrepository"
	ctrl "sigs.k8s.io/controller-runtime"
)

// choose right scenario to start (installation/deletion)
func sFnInitialize(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	// cleanup orphan resources
	if err := gitrepository.Cleanup(ctx, r.client); err != nil {
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigurationErr,
			err,
		)
		return stopWithEventualError(err)
	}

	// in case instance is being deleted and has finalizer - delete all resources
	instanceIsBeingDeleted := !s.instance.GetDeletionTimestamp().IsZero()
	if instanceIsBeingDeleted {
		return nextState(sFnDeleteResources)
	}

	return nextState(sFnRegistryConfiguration)
}
