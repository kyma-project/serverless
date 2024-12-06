package state

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

var requeueResult = &ctrl.Result{
	Requeue: true,
}

func nextState(next stateFn) (stateFn, *ctrl.Result, error) {
	return next, nil, nil
}

func requeue() (stateFn, *ctrl.Result, error) {
	return nil, requeueResult, nil
}

func requeueAfter(duration time.Duration) (stateFn, *ctrl.Result, error) {
	return nil, &ctrl.Result{
		RequeueAfter: duration,
	}, nil
}

func stop() (stateFn, *ctrl.Result, error) {
	return nil, nil, nil
}

func stopWithEventualError(err error) (stateFn, *ctrl.Result, error) {
	return nil, nil, err
}

func stopWithErrorOrRequeue(err error) (stateFn, *ctrl.Result, error) {
	return nil, requeueResult, err
}
