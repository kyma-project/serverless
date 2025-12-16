package state

import (
	"context"

	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	serverlessmetrics "github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/metrics"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnMetricsStart(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	f := m.State.Function
	serverlessmetrics.PublishReconciliationsTotal(f)
	serverlessmetrics.PublishFunctionsTotal(f)
	serverlessmetrics.StartForStateReachTime(f)

	return nextState(sFnCleanupLegacyServiceAccount)
}
