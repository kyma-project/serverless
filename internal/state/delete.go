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
func sFnDeleteResources() (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		// check if instance is in right state
		// TODO: in the future base it on conditions
		if s.instance.Status.State != v1alpha1.StateDeleting &&
			s.instance.Status.State != v1alpha1.StateError {
			return sFnUpdateServerlessStatus(v1alpha1.StateDeleting)
		}

		chartConfig, err := chartConfig(ctx, r, s)
		if err != nil {
			r.log.Errorf("error while preparing chart config: %s", err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		// TODO: thinkg about deletion configuration
		strategyStateFn := deletionStrategyBuilder(defaultDeletionStrategy)
		return strategyStateFn(chartConfig)
	}, nil, nil
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
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
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
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		// if resources are ready to be deleted, remove finalizer
		return sFnRemoveFinalizer(ctx, r, s)
	}, nil, nil
}
