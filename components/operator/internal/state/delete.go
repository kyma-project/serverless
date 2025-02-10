package state

import (
	"context"
	"time"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/chart"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultDeletionStrategy = safeDeletionStrategy
)

type deletionStrategy string

const (
	cascadeDeletionStrategy  deletionStrategy = "cascadeDeletionStrategy"
	safeDeletionStrategy     deletionStrategy = "safeDeletionStrategy"
	upstreamDeletionStrategy deletionStrategy = "upstreamDeletionStrategy"
)

// delete serverless based on previously installed resources
func sFnDeleteResources(_ context.Context, _ *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	s.setState(v1alpha1.StateDeleting)
	s.instance.UpdateConditionUnknown(
		v1alpha1.ConditionTypeDeleted,
		v1alpha1.ConditionReasonDeletion,
		"Uninstalling",
	)

	// TODO: thinkg about deletion configuration
	return nextState(
		deletionStrategyBuilder(defaultDeletionStrategy),
	)
}

func deletionStrategyBuilder(strategy deletionStrategy) stateFn {
	switch strategy {
	case cascadeDeletionStrategy:
		return sFnCascadeDeletionState
	case upstreamDeletionStrategy:
		return sFnUpstreamDeletionState
	case safeDeletionStrategy:
		return sFnSafeDeletionState
	default:
		return deletionStrategyBuilder(safeDeletionStrategy)
	}
}

func sFnCascadeDeletionState(_ context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	return deleteResourcesWithFilter(r, s)
}

func sFnUpstreamDeletionState(_ context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	return deleteResourcesWithFilter(r, s, chart.WithoutCRDFilter)
}

func sFnSafeDeletionState(_ context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	if err := chart.CheckCRDOrphanResources(s.chartConfig); err != nil {
		// stop state machine with a warning and requeue reconciliation in 1min
		// warning state indicates that user intervention would fix it. Its not reconciliation error.
		s.setState(v1alpha1.StateWarning)
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeDeleted,
			v1alpha1.ConditionReasonDeletionErr,
			err,
		)
		return stopWithEventualError(err)
	}

	return deleteResourcesWithFilter(r, s)
}

func deleteResourcesWithFilter(r *reconciler, s *systemState, filterFuncs ...chart.FilterFunc) (stateFn, *ctrl.Result, error) {
	resourceTypesToUninstall := []string{"Secret", "ConfigMap"}

	for _, resourceType := range resourceTypesToUninstall {
		err, done := chart.UninstallResourcesByType(s.chartConfig, resourceType, filterFuncs...)
		if err != nil {
			return uninstallResourcesError(r, s, err)
		}
		if !done {
			return awaitingResourcesRemoval(s, resourceType)
		}
	}

	if err := chart.Uninstall(s.chartConfig, filterFuncs...); err != nil {
		return uninstallResourcesError(r, s, err)
	}

	s.setState(v1alpha1.StateDeleting)
	s.instance.UpdateConditionTrue(
		v1alpha1.ConditionTypeDeleted,
		v1alpha1.ConditionReasonDeleted,
		"Serverless module deleted",
	)

	// if resources are ready to be deleted, remove finalizer
	return nextState(sFnRemoveFinalizer)
}

func uninstallResourcesError(r *reconciler, s *systemState, err error) (stateFn, *ctrl.Result, error) {
	r.log.Warnf("error while uninstalling resource %s: %s",
		client.ObjectKeyFromObject(&s.instance), err.Error())
	s.setState(v1alpha1.StateError)
	s.instance.UpdateConditionFalse(
		v1alpha1.ConditionTypeDeleted,
		v1alpha1.ConditionReasonDeletionErr,
		err,
	)
	return stopWithEventualError(err)
}

func awaitingResourcesRemoval(s *systemState, resourceType string) (stateFn, *ctrl.Result, error) {
	s.setState(v1alpha1.StateDeleting)

	s.instance.UpdateConditionTrue(
		v1alpha1.ConditionTypeDeleted,
		v1alpha1.ConditionReasonDeletion,
		"Deleting "+resourceType)

	// wait one sec until ctrl-mngr remove finalizers from secrets
	return requeueAfter(time.Second)
}
