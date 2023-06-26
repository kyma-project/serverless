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

func sFnUpdateProcessingState(condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, msg string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateProcessing)
		s.instance.UpdateConditionUnknown(condition, reason, msg)

		return updateServerlessStatus(buildSFnEmitEvent(sFnRequeue(), nil, nil), ctx, r, s)
	}
}

func sFnUpdateProcessingTrueState(condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, msg string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateProcessing)
		s.instance.UpdateConditionTrue(condition, reason, msg)

		return updateServerlessStatus(buildSFnEmitEvent(sFnRequeue(), nil, nil), ctx, r, s)
	}
}

func sFnUpdateReadyState(condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, msg string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateReady)
		s.instance.UpdateConditionTrue(condition, reason, msg)

		return updateServerlessStatus(buildSFnEmitEvent(sFnStop(), nil, nil), ctx, r, s)
	}
}

func sFnUpdateErrorState(condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, err error) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(condition, reason, err)

		return updateServerlessStatus(buildSFnEmitEvent(nil, nil, err), ctx, r, s)
	}
}

func sFnUpdateDeletingState(condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, msg string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateDeleting)
		s.instance.UpdateConditionUnknown(condition, reason, msg)

		return updateServerlessStatus(buildSFnEmitEvent(sFnRequeue(), nil, nil), ctx, r, s)
	}
}

func sFnUpdateDeletingTrueState(condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, msg string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateDeleting)
		s.instance.UpdateConditionTrue(condition, reason, msg)

		return updateServerlessStatus(buildSFnEmitEvent(sFnRequeue(), nil, nil), ctx, r, s)
	}
}

func sFnUpdateServerless() stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		return stopWithError(r.client.Update(ctx, &s.instance))
	}
}

func sFnUpdateServedTrue() stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setServed(v1alpha1.ServedTrue)
		return updateServerlessStatus(sFnRequeue(), ctx, r, s)
	}
}

func sFnUpdateWarningState(condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, msg string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateWarning)
		s.instance.UpdateConditionTrue(condition, reason, msg)

		return updateServerlessStatus(buildSFnEmitEvent(sFnStop(), nil, nil), ctx, r, s)
	}

}

func sFnUpdateServedFalse(condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, err error) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setServed(v1alpha1.ServedFalse)
		return nextState(sFnUpdateErrorState(condition, reason, err))
	}
}

func updateServerlessStatus(next stateFn, ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	instance := s.instance.DeepCopy()
	err := r.client.Status().Update(ctx, instance)
	if err != nil {
		return stopWithError(err)
	}
	return nextState(next)
}
