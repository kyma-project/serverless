package state

import (
	"context"

	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/metrics"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnConfigurationReady(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	// warn users when runtime is not supported
	msg := "Function configured"
	if !m.State.Function.Spec.Runtime.IsRuntimeSupported() {
		msg = "Warning: Function configured, runtime too old, used the latest supported runtime version"
	}

	m.State.Function.UpdateCondition(
		serverlessv1alpha2.ConditionConfigurationReady,
		metav1.ConditionTrue,
		serverlessv1alpha2.ConditionReasonFunctionSpecValidated,
		msg)
	metrics.PublishStateReachTime(m.State.Function, serverlessv1alpha2.ConditionConfigurationReady)

	return nextState(sFnHandleDeployment)
}
