package state

import (
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
	return nil, requeueResult(), nil
}

func requeueResult() *ctrl.Result {
	return &ctrl.Result{
		Requeue: true,
	}
}

func requeueAfter(duration time.Duration) (stateFn, *ctrl.Result, error) {
	return nil, &ctrl.Result{
		RequeueAfter: duration,
	}, nil
}
