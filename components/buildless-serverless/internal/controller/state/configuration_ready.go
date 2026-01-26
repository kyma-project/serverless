package state

import (
	"context"
	"fmt"

	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/metrics"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	configurationReadyMessage = "Function configured"
	warningConfigurationReady = "Warning: Function configured, runtime too old, used the latest supported runtime version"
)

func sFnConfigurationReady(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	condition := metav1.ConditionTrue
	// warn users when runtime is not supported
	msg := configurationReadyMessage
	reason := serverlessv1alpha2.ConditionReasonFunctionSpecValidated
	if !m.State.Function.Spec.Runtime.IsRuntimeSupported() {
		msg = fmt.Sprintf("Warning: invalid runtime value: cannot find runtime %s, using %s instead", m.State.Function.Spec.Runtime, m.State.Function.Spec.Runtime.SupportedRuntimeEquivalent())
		condition = metav1.ConditionFalse
		reason = serverlessv1alpha2.ConditionReasonFunctionSpecRuntimeOutdated
	}

	m.State.Function.UpdateCondition(
		serverlessv1alpha2.ConditionConfigurationReady,
		condition,
		reason,
		msg)
	metrics.PublishStateReachTime(m.State.Function, serverlessv1alpha2.ConditionConfigurationReady)

	return nextState(sFnHandleDeployment)
}
