package state

import (
	"context"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func sFnStart(ctx context.Context, m *stateMachine) (stateFn, *controllerruntime.Result, error) {
	return sFnHandleDeployment, &controllerruntime.Result{}, nil
}

func sFnFinish(ctx context.Context, m *stateMachine) (stateFn, *controllerruntime.Result, error) {
	return nil, &controllerruntime.Result{}, nil
}
