package state

import (
	"context"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// choose right scenario to start (installation/deletion)
func sFnInitialize(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	instanceIsBeingDeleted := !s.instance.GetDeletionTimestamp().IsZero()
	instanceHasFinalizer := controllerutil.ContainsFinalizer(&s.instance, r.finalizer)

	// in case instance does not have finalizer - add it and update instance
	if !instanceIsBeingDeleted && !instanceHasFinalizer {
		controllerutil.AddFinalizer(&s.instance, r.finalizer)
		err := r.client.Update(ctx, &s.instance)
		// stop state machine with potential error
		return stopWithError(err)
	}

	// in case instance has no finalizer and instance is being deleted - end reconciliation
	if instanceIsBeingDeleted && !instanceHasFinalizer {
		// stop state machine
		return stop()
	}

	if s.instance.Status.State.IsEmpty() {
		return sFnUpdateServerlessStatus(v1alpha1.StateProcessing)
	}

	err := s.Setup(ctx, r.client)
	if err != nil {
		return sFnUpdateServerlessStatus(v1alpha1.StateError)
	}

	// in case instance is being deleted and has finalizer - delete all resources
	if instanceIsBeingDeleted {
		return sFnDeleteResources()
	}

	return sFnPrerequisites()
}
