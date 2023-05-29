package state

import (
	"context"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func sFnRemoveFinalizer() stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		if controllerutil.RemoveFinalizer(&s.instance, r.finalizer) {
			return nextState(
				sFnUpdateServerless(),
			)
		}

		return nextState(
			sFnUpdateDeletingTrueState(
				v1alpha1.ConditionTypeDeleted,
				v1alpha1.ConditionReasonDeleted,
				"Serverless module deleted",
			),
		)
	}
}
