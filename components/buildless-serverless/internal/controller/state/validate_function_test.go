package state

import (
	"context"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/controller/fsm"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_sFnValidateFunction(t *testing.T) {
	t.Run("when function is valid should go to the next state", func(t *testing.T) {
		// Arrange
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "elated-turing-name",
						Namespace: "mystifying-snyder-ns"}}}}

		// Act
		next, result, err := sFnValidateFunction(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// without stopping processing
		require.Nil(t, result)
		// with expected next state
		require.NotNil(t, next)
		requireEqualFunc(t, sFnHandleGitSources, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionTrue,
			serverlessv1alpha2.ConditionReasonFunctionSpecValidated,
			"function spec validated")
	})
	t.Run("when function is invalid should stop processing", func(t *testing.T) {
		// Arrange
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "elated-turing-name",
						Namespace: "mystifying-snyder-ns"},
					Spec: serverlessv1alpha2.FunctionSpec{
						Runtime: "gracious-bardeen"}}}}

		// Act
		next, result, err := sFnValidateFunction(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// no result because of stop
		require.Nil(t, result)
		// no next state (we will stop)
		require.Nil(t, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonInvalidFunctionSpec,
			"invalid runtime value: cannot find runtime: gracious-bardeen")
	})
}
