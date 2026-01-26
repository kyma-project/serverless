package state

import (
	"context"
	"testing"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_sFnConfigurationReady(t *testing.T) {
	t.Run("should set condition and go to the next state", func(t *testing.T) {
		// Arrange
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					Spec: serverlessv1alpha2.FunctionSpec{
						Runtime: serverlessv1alpha2.NodeJs22,
					},
				},
			},
		}

		// Act
		next, result, err := sFnConfigurationReady(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// without stopping processing
		require.Nil(t, result)
		// with expected next state
		require.NotNil(t, next)
		requireEqualFunc(t, sFnHandleDeployment, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionTrue,
			serverlessv1alpha2.ConditionReasonFunctionSpecValidated,
			"Function configured")
	})
	t.Run("should set condition to false on old runtime and go to the next state", func(t *testing.T) {
		// Arrange
		// machine with our function
		m := fsm.StateMachine{State: fsm.SystemState{
			Function: serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs14,
				},
			},
		}}

		// Act
		next, result, err := sFnConfigurationReady(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// without stopping processing
		require.Nil(t, result)
		// with expected next state
		require.NotNil(t, next)
		requireEqualFunc(t, sFnHandleDeployment, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonFunctionSpecRuntimeFallback,
			"Warning: invalid runtime value: cannot find runtime nodejs14, using runtime nodejs20 as a fallback to migrate from legacy serverless")
	})
}
