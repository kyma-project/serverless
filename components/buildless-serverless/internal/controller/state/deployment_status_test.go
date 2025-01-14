package state

import (
	"context"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/config"
	"github.com/kyma-project/serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/internal/controller/resources"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func Test_sFnDeploymentStatus(t *testing.T) {
	t.Run("when deployment is ready should requeue after long time from config", func(t *testing.T) {
		// Arrange
		// deployment which will be returned from kubernetes
		deployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "strange-chebyshev-name",
				Namespace: "busy-ramanujan-ns"},
			Status: appsv1.DeploymentStatus{
				Conditions: []appsv1.DeploymentCondition{
					{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue, Reason: MinimumReplicasAvailable},
					{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue, Reason: NewRSAvailableReason}}}}
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, appsv1.AddToScheme(scheme))
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&deployment).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Deployment: &resources.Deployment{
					Deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "strange-chebyshev-name",
							Namespace: "busy-ramanujan-ns"}}},
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "strange-chebyshev-name",
						Namespace: "busy-ramanujan-ns"}}},
			FunctionConfig: config.FunctionConfig{
				FunctionReadyRequeueDuration: 3456},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnDeploymentStatus(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// we expect stop and requeue
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: 3456}, *result)
		// no next state (we will stop)
		require.Nil(t, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionTrue,
			serverlessv1alpha2.ConditionReasonDeploymentReady,
			"Deployment strange-chebyshev-name is ready")
	})
	t.Run("when deployment is unhealthy should requeue", func(t *testing.T) {
		// Arrange
		// deployment which will be returned from kubernetes
		deployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "peaceful-rhodes-name",
				Namespace: "eloquent-shockley-ns"},
			Status: appsv1.DeploymentStatus{
				Conditions: []appsv1.DeploymentCondition{
					{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionFalse, Reason: MinimumReplicasUnavailable}}}}
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, appsv1.AddToScheme(scheme))
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&deployment).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Deployment: &resources.Deployment{
					Deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "peaceful-rhodes-name",
							Namespace: "eloquent-shockley-ns"}}},
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "peaceful-rhodes-name",
						Namespace: "eloquent-shockley-ns"}}},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnDeploymentStatus(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// we expect stop and requeue
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: defaultRequeueTime}, *result)
		// no next state (we will stop)
		require.Nil(t, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionUnknown,
			serverlessv1alpha2.ConditionReasonMinReplicasNotAvailable,
			"Minimum replicas not available for deployment peaceful-rhodes-name")
	})
	t.Run("when deployment is not ready should requeue", func(t *testing.T) {
		// Arrange
		// deployment which will be returned from kubernetes
		deployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "eloquent-stonebraker-name",
				Namespace: "clever-diffie-ns"},
			Status: appsv1.DeploymentStatus{
				Conditions: []appsv1.DeploymentCondition{
					{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue}}}}
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, appsv1.AddToScheme(scheme))
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&deployment).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Deployment: &resources.Deployment{
					Deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "eloquent-stonebraker-name",
							Namespace: "clever-diffie-ns"}}},
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eloquent-stonebraker-name",
						Namespace: "clever-diffie-ns"}}},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnDeploymentStatus(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// we expect stop and requeue
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: defaultRequeueTime}, *result)
		// no next state (we will stop)
		require.Nil(t, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionUnknown,
			serverlessv1alpha2.ConditionReasonDeploymentWaiting,
			"Deployment eloquent-stonebraker-name is not ready yet")
	})
	t.Run("when deployment is ready should requeue after long time from config", func(t *testing.T) {
		// Arrange
		// deployment which will be returned from kubernetes
		deployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "strange-chebyshev-name",
				Namespace: "busy-ramanujan-ns"},
			Status: appsv1.DeploymentStatus{
				Conditions: []appsv1.DeploymentCondition{
					{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue, Reason: MinimumReplicasAvailable},
					{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue, Reason: NewRSAvailableReason}}}}
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, appsv1.AddToScheme(scheme))
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&deployment).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Deployment: &resources.Deployment{
					Deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "strange-chebyshev-name",
							Namespace: "busy-ramanujan-ns"}}},
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "strange-chebyshev-name",
						Namespace: "busy-ramanujan-ns"}}},
			FunctionConfig: config.FunctionConfig{
				FunctionReadyRequeueDuration: 3456},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnDeploymentStatus(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// we expect stop and requeue
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: 3456}, *result)
		// no next state (we will stop)
		require.Nil(t, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionTrue,
			serverlessv1alpha2.ConditionReasonDeploymentReady,
			"Deployment strange-chebyshev-name is ready")
	})
	t.Run("when deployment failed should stop processing", func(t *testing.T) {
		// Arrange
		// deployment which will be returned from kubernetes
		deployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "infallible-carver-name",
				Namespace: "jolly-galileo-ns"},
			Status: appsv1.DeploymentStatus{
				Conditions: []appsv1.DeploymentCondition{
					{Type: "fervent-engelbart", Status: corev1.ConditionTrue},
					{Type: "admiring-lovelace", Status: corev1.ConditionTrue}}}}
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, appsv1.AddToScheme(scheme))
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&deployment).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Deployment: &resources.Deployment{
					Deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "infallible-carver-name",
							Namespace: "jolly-galileo-ns"}}},
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "infallible-carver-name",
						Namespace: "jolly-galileo-ns"}}},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnDeploymentStatus(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// we expect stop without requeue
		require.Nil(t, result)
		// no next state (we will stop)
		require.Nil(t, next)
		// function has proper condition
		yamlConditions, _ := yaml.Marshal(deployment.Status.Conditions)
		expectedMsg := fmt.Sprintf("Deployment infallible-carver-name failed with condition: \n%s", yamlConditions)
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonDeploymentFailed,
			expectedMsg)
	})
}
