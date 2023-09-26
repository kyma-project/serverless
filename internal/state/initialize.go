package state

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

// choose right scenario to start (installation/deletion)
func sFnInitialize(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	// default instance and create necessary essentials
	s.setup(ctx, r)

	// in case instance is being deleted and has finalizer - delete all resources
	instanceIsBeingDeleted := !s.instance.GetDeletionTimestamp().IsZero()
	if instanceIsBeingDeleted {
		return nextState(sFnDeleteResources)
	}

	return nextState(sFnRegistryConfiguration)
}
