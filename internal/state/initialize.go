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
		return noFinalizerStep(ctx, r, s, instanceIsBeingDeleted)
	}

	// default instance and create necessary essentials
	s.setup(ctx, r)

	// in case instance is being deleted and has finalizer - delete all resources
	if instanceIsBeingDeleted {
		return nextState(sFnDeleteResources)
	}

	return nextState(sFnRegistryConfiguration)
}

func noFinalizerStep(ctx context.Context, r *reconciler, s *systemState, instanceIsBeingDeleted bool) (stateFn, *ctrl.Result, error) {
	// in case instance has no finalizer and instance is being deleted - end reconciliation
	if instanceIsBeingDeleted {
		// stop state machine
		return stop()
	}

	// in case instance does not have finalizer - add it and update instance
	controllerutil.AddFinalizer(&s.instance, r.finalizer)
	err := updateServerlessBody(ctx, r, s)
	// TODO: there is no need to requeue
	// stop state machine with potential error
	return stopWithErrorOrRequeue(err)
}
