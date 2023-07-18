package state

import (
	"context"
	"fmt"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// verify if all workloads are in ready state
func sFnVerifyResources(_ context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	ready, err := chart.Verify(s.chartConfig)
	if err != nil {
		r.log.Warnf("error while verifying resource %s: %s",
			client.ObjectKeyFromObject(&s.instance), err.Error())
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonInstallationErr,
			err,
		)
		return nextState(sFnUpdateStatusWithError(err))
	}

	if !ready {
		return requeueAfter(requeueDuration)
	}

	if s.warning != nil {
		s.setState(v1alpha1.StateWarning)
		s.instance.UpdateConditionTrue(
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonInstalled,
			fmt.Sprintf("%s: %s", warningMessagePrefix, s.warning.Error()),
		)
		return nextState(sFnUpdateStatusAndStop)
	}

	s.setState(v1alpha1.StateReady)
	s.instance.UpdateConditionTrue(
		v1alpha1.ConditionTypeInstalled,
		v1alpha1.ConditionReasonInstalled,
		"Serverless installed",
	)
	return nextState(sFnUpdateStatusAndStop)
}
