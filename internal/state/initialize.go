package state

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// choose right scenario to start (installation/deletion)
func sFnInitialize(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	instanceIsBeingDeleted := !s.instance.GetDeletionTimestamp().IsZero()
	instanceHasFinalizer := controllerutil.ContainsFinalizer(&s.instance, r.finalizer)
	if !instanceHasFinalizer {
		// in case instance has no finalizer and instance is being deleted - end reconciliation
		if instanceIsBeingDeleted {
			// stop state machine
			return stop()
		}

		if err := addFinalizer(ctx, r, s); err != nil {
			// stop state machine with potential error
			return stopWithPossibleError(err)
		}
	}

	// default instance and create necessary essentials
	s.setup(ctx, r)

	// in case instance is being deleted and has finalizer - delete all resources
	if instanceIsBeingDeleted {
		return nextState(sFnDeleteResources)
	}

	return nextState(sFnRegistryConfiguration)
}

func addFinalizer(ctx context.Context, r *reconciler, s *systemState) error {
	// in case instance does not have finalizer - add it and update instance
	controllerutil.AddFinalizer(&s.instance, r.finalizer)
	return updateServerlessWithoutStatus(ctx, r, s)
}
