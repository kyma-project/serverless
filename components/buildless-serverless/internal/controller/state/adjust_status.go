package state

import (
	"context"
	deployment_builder "github.com/kyma-project/serverless/internal/controller/deployment"
	"github.com/kyma-project/serverless/internal/controller/fsm"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnAdjustStatus(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	//TODO: Move set statuses to the final state
	m.State.Instance.Status.RuntimeImage = deployment_builder.New(m).RuntimeImage()
	//TODO: Add more status fields
	return nextState(sFnDeploymentStatus)
}
