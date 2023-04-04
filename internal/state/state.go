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

// TODO: remove me pls
func sFnUpdateServerlessStatus(next stateFn) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		err := r.client.Status().Update(ctx, &s.instance)
		if err != nil {
			stopWithError(err)
		}
		return next, nil, nil
	}, nil, nil
}

func sFnUpdateProcessingState(next stateFn, condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, msg string) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateProcessing)
		s.instance.UpdateConditionUnknown(condition, reason, msg)

		return sFnUpdateServerlessStatus(sFnEmmitEventfunc(next, nil, nil))
	}, nil, nil
}

func sFnUpdateProcessingTrueState(next stateFn, condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, msg string) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateProcessing)
		s.instance.UpdateConditionTrue(condition, reason, msg)

		return sFnUpdateServerlessStatus(sFnEmmitEventfunc(next, nil, nil))
	}, nil, nil
}

func sFnUpdateReadyState(next stateFn, condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, msg string) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateReady)
		s.instance.UpdateConditionTrue(condition, reason, msg)

		return sFnUpdateServerlessStatus(sFnEmmitEventfunc(next, nil, nil))
	}, nil, nil
}

func sFnUpdateErrorState(next stateFn, condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, err error) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(condition, reason, err)

		return sFnUpdateServerlessStatus(sFnEmmitEventfunc(next, nil, nil))
	}, nil, nil
}

func sFnUpdateDeletingState(next stateFn, condition v1alpha1.ConditionType, reason v1alpha1.ConditionReason, msg string) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(v1alpha1.StateDeleting)
		s.instance.UpdateConditionUnknown(condition, reason, msg)

		return sFnUpdateServerlessStatus(sFnEmmitEventfunc(next, nil, nil))
	}, nil, nil
}

func sFnUpdateServerless() (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		return nil, nil, r.client.Update(ctx, &s.instance)
	}, nil, nil
}

func stopWithErrorOrRequeue(err error) (stateFn, *ctrl.Result, error) {
	return nil, &ctrl.Result{
		Requeue: true,
	}, err
}

func stopWithError(err error) (stateFn, *ctrl.Result, error) {
	return nil, nil, err
}

func stop() (stateFn, *ctrl.Result, error) {
	return nil, nil, nil
}

func requeue() (stateFn, *ctrl.Result, error) {
	return nil, &ctrl.Result{
		Requeue: true,
	}, nil
}

func sFnStop() stateFn {
	return func(_ context.Context, _ *reconciler, _ *systemState) (stateFn, *ctrl.Result, error) {
		return stop()
	}
}

func sFnRequeue() stateFn {
	return func(_ context.Context, _ *reconciler, _ *systemState) (stateFn, *ctrl.Result, error) {
		return requeue()
	}
}

func requeueAfter(duration time.Duration) (stateFn, *ctrl.Result, error) {
	return nil, &ctrl.Result{
		RequeueAfter: duration,
	}, nil
}
