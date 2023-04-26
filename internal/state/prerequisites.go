package state

import (
	"context"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/dependencies"
	"k8s.io/apimachinery/pkg/api/meta"
	ctrl "sigs.k8s.io/controller-runtime"
)

// check necessery dependencies before installation
func buildSFnPrerequisites(s *systemState) stateFn {
	next := sFnPrerequisites
	if meta.FindStatusCondition(s.instance.Status.Conditions, string(v1alpha1.ConditionTypeConfigured)) == nil {
		next = sFnUpdateProcessingState(
			sFnPrerequisites,
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonPrerequisites,
			"Checking prerequisites",
		)
	}
	return next
}

func sFnPrerequisites(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	// check hard serverless dependencies before installation
	withIstio := s.instance.Spec.DockerRegistry.IsInternalEnabled()
	err := dependencies.CheckPrerequisites(ctx, r.client, withIstio)
	if err != nil {
		return nextState(
			sFnUpdateErrorState(
				sFnRequeue(),
				v1alpha1.ConditionTypeConfigured,
				v1alpha1.ConditionReasonPrerequisitesErr,
				err,
			),
		)
	}

	// when we know that cluster configuration met serverless requirements we can go to installation state
	return nextState(
		sFnUpdateProcessingTrueState(
			buildSFnApplyResources(s),
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonPrerequisitesMet,
			"All prerequisites met",
		),
	)
}
