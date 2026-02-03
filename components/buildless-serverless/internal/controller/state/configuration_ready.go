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
	configurationReadyMessage      = "Function configured"
	warningDeprecatedRuntimeFormat = "Warning: function configured, runtime %s is deprecated and will be removed in the future"
)

func sFnConfigurationReady(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	msg := configurationReadyMessage
	reason := serverlessv1alpha2.ConditionReasonFunctionSpecValidated

	if m.State.Function.Spec.Runtime.IsRuntimeDeprecated() {
		// warn users when runtime is deprecated
		msg = fmt.Sprintf(warningDeprecatedRuntimeFormat, m.State.Function.Spec.Runtime)
	}

	m.State.Function.UpdateCondition(
		serverlessv1alpha2.ConditionConfigurationReady,
		metav1.ConditionTrue,
		reason,
		msg)
	metrics.PublishStateReachTime(m.State.Function, serverlessv1alpha2.ConditionConfigurationReady)

	return nextState(sFnHandleDeployment)
}
