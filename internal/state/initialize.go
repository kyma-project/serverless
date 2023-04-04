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
		return stopWithErrorOrRequeue(err)
	}

	// in case instance has no finalizer and instance is being deleted - end reconciliation
	if instanceIsBeingDeleted && !instanceHasFinalizer {
		// stop state machine
		return stop()
	}

	err := s.Setup(ctx, r)
	if err != nil {
		return sFnUpdateErrorState(
			sFnRequeue(),
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonPrerequisitesErr,
			err,
		)
	}

	// in case instance is being deleted and has finalizer - delete all resources
	if instanceIsBeingDeleted {
		return buildSFnDeleteResources()
	}

	return buildSFnPrerequisites(s)
}
