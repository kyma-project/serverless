package state

import (
	"context"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnAdjustStatus(_ context.Context, m *stateMachine) (stateFn, *ctrl.Result, error) {
	m.state.instance.Status.RuntimeImage = m.getRuntimeImage()
	return sFnDeploymentStatus, nil, nil
}
