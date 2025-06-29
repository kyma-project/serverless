package state

import (
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

var requeueResult = &ctrl.Result{
	Requeue: true,
}

func nextState(next fsm.StateFn) (fsm.StateFn, *ctrl.Result, error) {
	return next, nil, nil
}

func requeue() (fsm.StateFn, *ctrl.Result, error) {
	return nil, requeueResult, nil
}

func requeueAfter(duration time.Duration) (fsm.StateFn, *ctrl.Result, error) {
	return nil, &ctrl.Result{
		RequeueAfter: duration,
	}, nil
}

func requeueAfterWithError(duration time.Duration, err error) (fsm.StateFn, *ctrl.Result, error) {
	return nil, &ctrl.Result{
		RequeueAfter: duration,
	}, err
}

func stop() (fsm.StateFn, *ctrl.Result, error) {
	return nil, nil, nil
}

func stopWithError(err error) (fsm.StateFn, *ctrl.Result, error) {
	return nil, nil, err
}

func StartState() fsm.StateFn {
	return sFnMetricsStart
}
