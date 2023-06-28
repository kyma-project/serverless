package state

import (
	"context"
	"time"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	requeueDuration = time.Second * 3
)

func sFnUpdateStatusAndRequeue(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	err := updateServerlessStatus(ctx, r, s)
	return sFnRequeue(), nil, err
}

func sFnUpdateStatusAndStop(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	err := updateServerlessStatus(ctx, r, s)
	return sFnStop(), nil, err
}

func sFnUpdateErrorState(condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, err error) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(condition, reason, err)

		err := updateServerlessStatus(ctx, r, s)
		return nil, nil, err
	}
}

func sFnUpdateDeletingTrueState(condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, msg string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateDeleting)
		s.instance.UpdateConditionTrue(condition, reason, msg)

		err := updateServerlessStatus(ctx, r, s)
		return sFnRequeue(), nil, err
	}
}

func sFnUpdateServerless() stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		return stopWithError(r.client.Update(ctx, &s.instance))
	}
}

func sFnUpdateWarningState(condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, msg string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateWarning)
		s.instance.UpdateConditionTrue(condition, reason, msg)
		err := updateServerlessStatus(ctx, r, s)
		return sFnStop(), nil, err
	}

}

func updateServerlessStatus(ctx context.Context, r *reconciler, s *systemState) error {
	instance := s.instance.DeepCopy()
	err := r.client.Status().Update(ctx, instance)
	emitEvent(r, s)
	return err
}
