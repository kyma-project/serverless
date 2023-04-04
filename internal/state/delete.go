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
func buildSFnDeleteResources() (stateFn, *ctrl.Result, error) {
	return sFnUpdateDeletingState(
		sFnDeleteResources,
		v1alpha1.ConditionTypeInstalled,
		v1alpha1.ConditionReasonDeletion,
		"Uninstalling",
	)
}

func sFnDeleteResources(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	// TODO: thinkg about deletion configuration
	strategyStateFn := deletionStrategyBuilder(defaultDeletionStrategy)
	return strategyStateFn(s.chartConfig)
}

type deletionStrategyBuilderFn func(*chart.Config) (stateFn, *ctrl.Result, error)

func deletionStrategyBuilder(strategy deletionStrategy) deletionStrategyBuilderFn {
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

func sFnCascadeDeletionState(chartConfig *chart.Config) (stateFn, *ctrl.Result, error) {
	return deleteResourcesWithFilter(chartConfig)
}

func sFnUpstreamDeletionState(chartConfig *chart.Config) (stateFn, *ctrl.Result, error) {
	return deleteResourcesWithFilter(chartConfig, chart.WithoutCRDFilter)
}

func sFnSafeDeletionState(chartConfig *chart.Config) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, ss *systemState) (stateFn, *ctrl.Result, error) {
		if err := chart.CheckCRDOrphanResources(chartConfig); err != nil {
			// stop state machine with an error and requeue reconciliation in 1min
			return sFnUpdateErrorState(
				sFnRequeue(),
				v1alpha1.ConditionTypeInstalled,
				v1alpha1.ConditionReasonDeletionErr,
				err,
			)
		}

		return deleteResourcesWithFilter(chartConfig)
	}, nil, nil
}

func deleteResourcesWithFilter(chartConfig *chart.Config, filterFuncs ...chart.FilterFunc) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		err := chart.Uninstall(chartConfig)
		if err != nil {
			r.log.Warnf("error while uninstalling resource %s: %s",
				client.ObjectKeyFromObject(&s.instance), err.Error())
			return sFnUpdateErrorState(
				sFnRequeue(),
				v1alpha1.ConditionTypeInstalled,
				v1alpha1.ConditionReasonDeletionErr,
				err,
			)
		}

		// if resources are ready to be deleted, remove finalizer
		return sFnUpdateReadyState(
			sFnRemoveFinalizer(),
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonDeleted,
			"Serverless module deleted",
		)

	}, nil, nil
}
