package state

import (
	"context"
	"testing"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnDeleteDeployments(t *testing.T) {
	t.Run("should delete labeled deployments", func(t *testing.T) {
		// Arrange
		// our function
		f := serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bold-galois",
				Namespace: "youthful-brahmagupta"},
			Spec: serverlessv1alpha2.FunctionSpec{
				Runtime: serverlessv1alpha2.NodeJs22,
				Source: serverlessv1alpha2.Source{
					Inline: &serverlessv1alpha2.InlineSource{
						Source: "confident-ardinghelli"}},
				Annotations: map[string]string{"joliot": "condescending"}}}
		// some deployment on k8s, but it is not the deployment we expect
		someDeployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "peaceful-leavitt",
				Namespace: f.GetNamespace()}}
		// two deployments on k8s with labels we expect
		firstDeployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "loving-aryabhata",
				Namespace: f.GetNamespace(),
				Labels:    f.InternalFunctionLabels()}}
		secondDeployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kind-germain",
				Namespace: f.GetNamespace(),
				Labels:    f.InternalFunctionLabels()}}
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, appsv1.AddToScheme(scheme))
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).
			WithObjects(&someDeployment, &firstDeployment, &secondDeployment).Build()
		// machine with our function
		m := fsm.StateMachine{
			State:  fsm.SystemState{Function: f},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnDeleteDeployments(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// we expect stop and requeue
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{Requeue: true}, *result)
		// no next state (we will stop)
		require.Nil(t, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonDeploymentDeleted,
			"Duplicated Deployments deleted")
		// deployment has been applied to k8s
		clusterDeployments := &appsv1.DeploymentList{}
		getErr := k8sClient.List(context.Background(), clusterDeployments)
		require.NoError(t, getErr)
		// labeled deployments should be deleted
		require.Len(t, clusterDeployments.Items, 1)
		require.Equal(t, "peaceful-leavitt", clusterDeployments.Items[0].Name)
	})
}
