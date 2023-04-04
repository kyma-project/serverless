package state

import (
	"context"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// verify if all workloads are in ready state
func sFnVerifyResources() (stateFn, *ctrl.Result, error) {
	return verifyResources, nil, nil
}

func verifyResources(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	ready, err := chart.Verify(s.chartConfig)
	if err != nil {
		r.log.Warnf("error while verifying resource %s: %s",
			client.ObjectKeyFromObject(&s.instance), err.Error())
		return sFnUpdateErrorState(
			sFnRequeue(),
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonInstallationErr,
			err,
		)
	}

	if ready {
		return sFnUpdateReadyState(
			sFnStop(),
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonInstalled,
			"Serverless installed",
		)
	}

	return requeueAfter(requeueDuration)
}
