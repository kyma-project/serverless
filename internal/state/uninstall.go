package state

import (
	"context"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func sFnDeleteResources() (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		// check if instance is in right state
		if s.instance.Status.State != v1alpha1.StateDeleting {
			return sFnUpdateServerlessStatus(v1alpha1.StateDeleting)
		}

		chartConfig, err := chartConfig(ctx, r, s)
		if err != nil {
			r.log.Errorf("error while preparing chart config: %s", err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		err = chart.Uninstall(chartConfig)
		if err != nil {
			r.log.Warnf("error while uninstalling resource %s: %s",
				client.ObjectKeyFromObject(&s.instance), err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		// if resources are ready to be deleted, remove finalizer
		if controllerutil.RemoveFinalizer(&s.instance, r.finalizer) {
			return sFnUpdateServerless()
		}

		return requeue()
	}, nil, nil
}
