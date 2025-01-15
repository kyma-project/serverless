package state

import (
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/internal/controller/resources"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func Test_sFnAdjustStatus(t *testing.T) {
	t.Run("status is set", func(t *testing.T) {
		// Arrange
		// machine with our function and previously created/calculated deployment
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					Status: serverlessv1alpha2.FunctionStatus{}},
				Deployment: &resources.Deployment{
					Deployment: &appsv1.Deployment{
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{Image: "zen-wu-image"}}}}}}}}}

		// Act
		next, result, err := sFnAdjustStatus(nil, &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// without stopping processing
		require.Nil(t, result)
		// with expected next state
		require.NotNil(t, next)
		requireEqualFunc(t, sFnDeploymentStatus, next)
		// function should have status image from deployment
		require.Equal(t, "zen-wu-image", m.State.Function.Status.RuntimeImage)
	})
}
