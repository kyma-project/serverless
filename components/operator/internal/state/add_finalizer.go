package state

import (
	"context"

	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func sFnAddFinalizer(ctx context.Context, r *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
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
			return stopWithEventualError(err)
		}
	}
	return nextState(sFnInitialize)
}

func addFinalizer(ctx context.Context, r *reconciler, s *systemState) error {
	// in case instance does not have finalizer - add it and update instance
	controllerutil.AddFinalizer(&s.instance, r.finalizer)
	return updateDockerRegistryWithoutStatus(ctx, r, s)
}
