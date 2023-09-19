package state

import (
	"context"
	"fmt"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnServedFilter(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	if s.instance.IsServedEmpty() {
		servedServerless, err := GetServedServerless(ctx, r.k8s.client)
		if err != nil {
			return stopWithError(err)
		}

		if servedServerless == nil {
			s.setServed(v1alpha1.ServedTrue)
			return nextState(sFnInitialize)
		}
		s.setServed(v1alpha1.ServedFalse)
		s.setState(v1alpha1.StateError)
		err = fmt.Errorf("only one instance of Serverless is allowed (current served instance: %s/%s)",
			servedServerless.GetNamespace(), servedServerless.GetName())
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonServerlessDuplicated,
			err,
		)
		return stopWithError(err)
	}

	if s.instance.Status.Served == v1alpha1.ServedFalse {
		return nil, nil, nil
	}

	return nextState(sFnInitialize)
}
