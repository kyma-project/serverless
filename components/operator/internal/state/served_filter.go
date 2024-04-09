package state

import (
	"context"
	"fmt"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnServedFilter(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	if s.instance.IsServedEmpty() {
		if err := setInitialServed(ctx, r, s); err != nil {
			return stopWithEventualError(err)
		}
	}

	if s.instance.Status.Served == v1alpha1.ServedFalse {
		return stop()
	}
	return nextState(sFnAddFinalizer)
}

func setInitialServed(ctx context.Context, r *reconciler, s *systemState) error {
	servedDockerRegistry, err := GetServedDockerRegistry(ctx, r.k8s.client)
	if err != nil {
		return err
	}

	return setServed(servedDockerRegistry, s)
}

func setServed(servedDockerRegistry *v1alpha1.DockerRegistry, s *systemState) error {
	if servedDockerRegistry == nil {
		s.setServed(v1alpha1.ServedTrue)
		return nil
	}

	s.setServed(v1alpha1.ServedFalse)
	s.setState(v1alpha1.StateWarning)
	err := fmt.Errorf(
		"Only one instance of DockerRegistry is allowed (current served instance: %s/%s). This DockerRegistry CR is redundant. Remove it to fix the problem.",
		servedDockerRegistry.GetNamespace(), servedDockerRegistry.GetName())
	s.instance.UpdateConditionFalse(
		v1alpha1.ConditionTypeConfigured,
		v1alpha1.ConditionReasonDuplicated,
		err,
	)
	return err
}
