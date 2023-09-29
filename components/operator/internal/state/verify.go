package state

import (
	"context"

	//TODO: change to "github.com/kyma-project/serverless-manager/components/operator/api/v1alpha1" after new SO structure will be on github
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	//TODO: change to "github.com/kyma-project/serverless-manager/components/operator/internal/chart" after new SO structure will be on github
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
		return stopWithEventualError(err)
	}

	if !ready {
		return requeueAfter(requeueDuration)
	}

	warning := s.warningBuilder.Build()
	if warning != "" {
		s.setState(v1alpha1.StateWarning)
		s.instance.UpdateConditionTrue(
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonInstalled,
			warning,
		)
		return stop()
	}

	s.setState(v1alpha1.StateReady)
	s.instance.UpdateConditionTrue(
		v1alpha1.ConditionTypeInstalled,
		v1alpha1.ConditionReasonInstalled,
		"Serverless installed",
	)
	return stop()
}
