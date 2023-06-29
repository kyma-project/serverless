package state

import (
	"context"
	"fmt"
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnServedFilter(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	if s.instance.IsServedEmpty() {
		servedServerless, err := findServedServerless(ctx, r.k8s.client)
		if err != nil {
			return stopWithError(err)
		}

		if servedServerless == nil {
			s.setServed(v1alpha1.ServedTrue)
			return nextState(sFnUpdateStatusAndRequeue)
		}
		s.setServed(v1alpha1.ServedFalse)
		setErrorState(s,
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonServerlessDuplicated,
			fmt.Errorf("only one instance of Serverless is allowed (current served instance: %s/%s)",
				servedServerless.GetNamespace(), servedServerless.GetName()),
		)
		return nextState(sFnUpdateStatusAndStop)
	}

	if s.instance.Status.Served == v1alpha1.ServedFalse {
		return nil, nil, nil
	}

	return nextState(sFnInitialize)
}

func findServedServerless(ctx context.Context, c client.Client) (*v1alpha1.Serverless, error) {
	var serverlessList v1alpha1.ServerlessList

	err := c.List(ctx, &serverlessList)

	if err != nil {
		return nil, err
	}

	for _, item := range serverlessList.Items {
		if !item.IsServedEmpty() && item.Status.Served == v1alpha1.ServedTrue {
			return &item, nil
		}
	}

	return nil, nil
}
