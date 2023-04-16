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

func sFnUpdateProcessingState(next stateFn, condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, msg string) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateProcessing)
		s.instance.UpdateConditionUnknown(condition, reason, msg)

		return updateServerlessStatus(sFnEmitEventfunc(next, nil, nil), ctx, r, s)
	}, nil, nil
}

func sFnUpdateProcessingTrueState(next stateFn, condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, msg string) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateProcessing)
		s.instance.UpdateConditionTrue(condition, reason, msg)

		return updateServerlessStatus(sFnEmitEventfunc(next, nil, nil), ctx, r, s)
	}, nil, nil
}

func sFnUpdateReadyState(next stateFn, condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, msg string) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateReady)
		s.instance.UpdateConditionTrue(condition, reason, msg)

		return updateServerlessStatus(sFnEmitEventfunc(next, nil, nil), ctx, r, s)
	}, nil, nil
}

func sFnUpdateErrorState(next stateFn, condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, err error) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(condition, reason, err)

		return updateServerlessStatus(sFnEmitEventfunc(next, nil, nil), ctx, r, s)
	}, nil, nil
}

func sFnUpdateDeletingState(next stateFn, eventType, eventReason, eventMessage string) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateDeleting)

		return updateServerlessStatus(sFnEmitStrictEvent(
			next, nil, nil,
			eventType,
			eventReason,
			eventMessage,
		), ctx, r, s)
	}, nil, nil
}

func sFnUpdateServerless() (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		return nil, nil, r.client.Update(ctx, &s.instance)
	}, nil, nil
}

func updateServerlessStatus(next stateFn, ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	err := r.client.Status().Update(ctx, &s.instance)
	if err != nil {
		stopWithError(err)
	}
	return next, nil, nil
}
