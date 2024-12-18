package state

import (
	"context"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/controller/fsm"
	functionvalidator "github.com/kyma-project/serverless/internal/controller/validator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
)

func sFnValidateFunction(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	result := functionvalidator.New(&m.State.Instance).Validate()

	if len(result) != 0 {
		//TODO: Use ConditionConfigure in this place
		m.State.Instance.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonFunctionSpec,
			strings.Join(result, ". "))
		return stop()
	}

	return nextState(sFnHandleDeployment)
}
