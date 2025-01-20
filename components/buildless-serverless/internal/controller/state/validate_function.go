package state

import (
	"context"
	"github.com/kyma-project/serverless/internal/controller/functionvalidator"
	"strings"

	"github.com/kyma-project/serverless/internal/controller/fsm"

	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnValidateFunction(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	v := functionvalidator.New(&m.State.Function)
	validationResults := v.Validate()
	if len(validationResults) != 0 {
		//TODO: Use ConditionConfigure in this place
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonFunctionSpec,
			strings.Join(validationResults, ". "))
		return stop()
	}

	return nextState(sFnHandleDeployment)
}
