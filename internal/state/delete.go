package state

import (
	"context"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
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
func sFnDeleteResources() stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		if !s.instance.IsCondition(v1alpha1.ConditionTypeDeleted) {
			return nextState(
				sFnUpdateDeletingState(
					v1alpha1.ConditionTypeDeleted,
					v1alpha1.ConditionReasonDeletion,
					"Uninstalling",
				),
			)
		}

		// TODO: thinkg about deletion configuration
		return nextState(
			deletionStrategyBuilder(defaultDeletionStrategy),
		)
	}
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
		// stop state machine with an error and requeue reconciliation in 1min
		return nextState(
			sFnUpdateErrorState(
				v1alpha1.ConditionTypeDeleted,
				v1alpha1.ConditionReasonDeletionErr,
				err,
			),
		)
	}

	return deleteResourcesWithFilter(r, s)
}

func deleteResourcesWithFilter(r *reconciler, s *systemState, filterFuncs ...chart.FilterFunc) (stateFn, *ctrl.Result, error) {
	err := chart.Uninstall(s.chartConfig, filterFuncs...)
	if err != nil {
		r.log.Warnf("error while uninstalling resource %s: %s",
			client.ObjectKeyFromObject(&s.instance), err.Error())
		return nextState(
			sFnUpdateErrorState(
				v1alpha1.ConditionTypeDeleted,
				v1alpha1.ConditionReasonDeletionErr,
				err,
			),
		)
	}

	// if resources are ready to be deleted, remove finalizer
	return nextState(
		sFnRemoveFinalizer(),
	)
}
