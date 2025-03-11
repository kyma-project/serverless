package state

import (
	"context"
	"github.com/kyma-project/serverless/internal/controller/fsm"

	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnConfigurationReady(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	m.State.Function.UpdateCondition(
		serverlessv1alpha2.ConditionConfigurationReady,
		metav1.ConditionTrue,
		serverlessv1alpha2.ConditionReasonFunctionSpecValidated,
		"function configured")
	return nextState(sFnHandleDeployment)
}
