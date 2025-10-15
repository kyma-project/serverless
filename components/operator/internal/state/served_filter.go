package state

import (
	"context"
	"fmt"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnServedFilter(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	if s.instance.IsServedEmpty() || s.instance.Status.Served == v1alpha1.ServedFalse {
		if err := calculateServed(ctx, r, s); err != nil {
			return stopWithEventualError(err)
		}
	}

	return nextState(sFnAddFinalizer)
}

func calculateServed(ctx context.Context, r *reconciler, s *systemState) error {
	servedServerless, err := GetServedServerless(ctx, r.k8s.client)
	if err != nil {
		return err
	}

	return setServed(servedServerless, s)
}

func setServed(servedServerless *v1alpha1.Serverless, s *systemState) error {
	if servedServerless == nil {
		s.setServed(v1alpha1.ServedTrue)
		return nil
	}

	s.setServed(v1alpha1.ServedFalse)
	s.setState(v1alpha1.StateWarning)
	err := fmt.Errorf(
		"only one instance of Serverless is allowed (current served instance: %s/%s) - this Serverless CR is redundant - remove it to fix the problem",
		servedServerless.GetNamespace(), servedServerless.GetName())
	s.instance.UpdateConditionFalse(
		v1alpha1.ConditionTypeConfigured,
		v1alpha1.ConditionReasonServerlessDuplicated,
		err,
	)
	return err
}
