package state

import (
	"context"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

func stopWithErrorOrRequeue(err error) (stateFn, *ctrl.Result, error) {
	return nil, &ctrl.Result{
		Requeue: true,
	}, err
}

func nextState(next stateFn) (stateFn, *ctrl.Result, error) {
	return next, nil, nil
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

func requeueAfter(duration time.Duration) (stateFn, *ctrl.Result, error) {
	return nil, &ctrl.Result{
		RequeueAfter: duration,
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
