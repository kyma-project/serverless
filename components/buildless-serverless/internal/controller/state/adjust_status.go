package state

import (
	"context"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func sFnAdjustStatus(_ context.Context, m *stateMachine) (stateFn, *controllerruntime.Result, error) {
	m.state.instance.Status.RuntimeImage = m.getRuntimeImage()
	return sFnFinish, nil, nil
}
