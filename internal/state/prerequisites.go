package state

import (
	"context"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/dependencies"
	ctrl "sigs.k8s.io/controller-runtime"
)

// check necessery dependencies before installation
func sFnPrerequisites() stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		// check if condition exists
		if !s.instance.IsCondition(v1alpha1.ConditionTypeConfigured) {
			return nextState(
				sFnUpdateProcessingState(
					v1alpha1.ConditionTypeConfigured,
					v1alpha1.ConditionReasonPrerequisites,
					"Checking prerequisites",
				),
			)
		}

		// check hard serverless dependencies before installation
		withIstio := s.instance.Spec.DockerRegistry.IsInternalEnabled()
		err := dependencies.CheckPrerequisites(ctx, r.client, withIstio)
		if err != nil {
			return nextState(
				sFnUpdateErrorState(
					v1alpha1.ConditionTypeConfigured,
					v1alpha1.ConditionReasonPrerequisitesErr,
					err,
				),
			)
		}

		// set condition before next state
		if !s.instance.IsConditionTrue(v1alpha1.ConditionTypeConfigured) {
			return nextState(
				sFnUpdateProcessingTrueState(
					v1alpha1.ConditionTypeConfigured,
					v1alpha1.ConditionReasonPrerequisitesMet,
					"All prerequisites met",
				),
			)
		}

		// when we know that cluster configuration met serverless requirements we can go to installation state
		return nextState(
			sFnApplyResources(),
		)
	}
}
