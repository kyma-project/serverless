package state

import (
	"context"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// run serverless chart installation
func sFnApplyResources() (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		chartConfig, err := chartConfig(ctx, r, s)
		if err != nil {
			r.log.Errorf("error while preparing chart config: %s", err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		err = chart.Install(chartConfig)
		if err != nil {
			r.log.Warnf("error while installing resource %s: %s",
				client.ObjectKeyFromObject(&s.instance), err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		return sFnVerifyResources(chartConfig)
	}, nil, nil
}
