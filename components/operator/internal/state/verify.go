package state

import (
	"context"
	"errors"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/chart"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// verify if all workloads are in ready state
func sFnVerifyResources(_ context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	result, err := chart.Verify(s.chartConfig)
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

	if !result.Ready && result.Reason == chart.DeploymentVerificationProcessing {
		// still processing
		return requeueAfter(requeueDuration)
	}

	if !result.Ready {
		// verification failed
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionTrue(
			v1alpha1.ConditionTypeDeploymentFailure,
			v1alpha1.ConditionReasonDeploymentReplicaFailure,
			result.Reason,
		)
		return stopWithEventualError(errors.New(result.Reason))
	}

	// remove possible previous DeploymentFailure condition
	s.instance.RemoveCondition(v1alpha1.ConditionTypeDeploymentFailure)

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

	if s.instance.Status.State != v1alpha1.StateReady {
		// set to Ready state if not already there
		s.setState(v1alpha1.StateReady)
		s.instance.UpdateConditionTrue(
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonInstalled,
			"Serverless installed",
		)
		// requeue to reconcile one more time and double check everything is fine
		return requeue()
	}

	// already ready
	return stop()
}
