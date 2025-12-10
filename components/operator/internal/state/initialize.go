package state

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

// choose right scenario to start (installation/deletion)
func sFnInitialize(_ context.Context, _ *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	// in case instance is being deleted and has finalizer - delete all resources
	instanceIsBeingDeleted := !s.instance.GetDeletionTimestamp().IsZero()
	if instanceIsBeingDeleted {
		return nextState(sFnDeleteResources)
	}

	//TODO: this is temporary solution, delete it after removing legacy serverless
	if isLegacyEnabled(s.instance.Annotations) {
		return nextState(sFnRegistryConfiguration)
	}
	return nextState(sFnOptionalDependencies)
}
