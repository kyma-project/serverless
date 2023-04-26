package state

import (
	"context"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// run serverless chart installation
func buildSFnApplyResources(s *systemState) stateFn {
	next := sFnApplyResources
	if !s.isCondition(v1alpha1.ConditionTypeInstalled) {
		next = sFnUpdateProcessingState(
			sFnApplyResources,
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonInstallation,
			"Installing for configuration",
		)
	}

	return next
}

func sFnApplyResources(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	err := chart.Install(s.chartConfig)
	if err != nil {
		r.log.Warnf("error while installing resource %s: %s",
			client.ObjectKeyFromObject(&s.instance), err.Error())
		return nextState(
			sFnUpdateErrorState(
				sFnRequeue(),
				v1alpha1.ConditionTypeInstalled,
				v1alpha1.ConditionReasonInstallationErr,
				err,
			),
		)
	}

	return nextState(
		buildSFnVerifyResources(),
	)
}
