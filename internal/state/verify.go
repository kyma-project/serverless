package state

import (
	"context"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// verify if all workloads are in ready state
func sFnVerifyResources(chartConfig *chart.Config) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		ready, err := chart.Verify(chartConfig)
		if err != nil {
			r.log.Warnf("error while verifying resource %s: %s",
				client.ObjectKeyFromObject(&s.instance), err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		if ready {
			return sFnUpdateServerlessStatus(v1alpha1.StateReady)
		}

		return requeueAfter(requeueDuration)
	}, nil, nil
}
