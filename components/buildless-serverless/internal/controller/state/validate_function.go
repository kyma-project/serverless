package state

import (
	"context"
	"strings"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/validator"
	"github.com/kyma-project/serverless/components/common/fips"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnValidateFunction(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	//TODO: It is a temporary solution to delete obsolete condition. It should be removed after migration from old serverless
	meta.RemoveStatusCondition(&m.State.Function.Status.Conditions, "BuildReady")

	v := validator.New(&m.State.Function, m.FunctionConfig, fips.IsFIPS140Only)
	validationResults := v.Validate()
	if len(validationResults) != 0 {
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonInvalidFunctionSpec,
			strings.Join(validationResults, ". "))
		return stop()
	}

	return nextState(sFnHandleGitSources)
}
