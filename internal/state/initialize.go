package state

import (
	"context"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// choose right scenario to start (installation/deletion)
var sFnInitialize = func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	s.saveServerlessStatus()

	instanceIsBeingDeleted := !s.instance.GetDeletionTimestamp().IsZero()
	instanceHasFinalizer := controllerutil.ContainsFinalizer(&s.instance, r.finalizer)
	if !instanceHasFinalizer {
		return noFinalizerStep(ctx, r, s, instanceIsBeingDeleted)
	}

	// default instance and create necessary essentials
	s.setup(ctx, r)

	// in case instance is being deleted and has finalizer - delete all resources
	if instanceIsBeingDeleted {
		return nextState(
			sFnDeleteResources(),
		)
	}

	err := s.setConfigFlags(ctx, r)
	if err != nil {
		return nextState(
			sFnUpdateErrorState(
				v1alpha1.ConditionTypeConfigured,
				v1alpha1.ConditionReasonConfigurationErr,
				err,
			),
		)
	}

	return nextState(
		sFnOptionalDependencies,
	)
}

func noFinalizerStep(ctx context.Context, r *reconciler, s *systemState, instanceIsBeingDeleted bool) (stateFn, *ctrl.Result, error) {
	// in case instance has no finalizer and instance is being deleted - end reconciliation
	if instanceIsBeingDeleted {
		// stop state machine
		return stop()
	}

	// in case instance does not have finalizer - add it and update instance
	controllerutil.AddFinalizer(&s.instance, r.finalizer)
	err := r.client.Update(ctx, &s.instance)
	// stop state machine with potential error
	return stopWithErrorOrRequeue(err)
}
