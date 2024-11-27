package state

import (
	"context"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func sFnDeploymentStatus(ctx context.Context, m *stateMachine) (stateFn, *ctrl.Result, error) {
	deploymentName := m.deploymentName()
	deployment := appsv1.Deployment{}
	m.client.Get(ctx, client.ObjectKey{
		Namespace: m.state.instance.Namespace,
		Name:      deploymentName,
	}, &deployment)

	// ready deployment
	if isDeploymentReady(deployment) {
		m.log.Info(fmt.Sprintf("deployment %s ready", deploymentName))

		m.state.instance.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionTrue,
			serverlessv1alpha2.ConditionReasonDeploymentReady,
			fmt.Sprintf("Deployment %s is ready", deploymentName))

		return sFnFinish /*&ctrl.Result{RequeueAfter: m.functionConfig.FunctionReadyRequeueDuration}*/, nil, nil
	}

	return sFnFinish, nil, nil
}

const (
	// Progressing:
	// NewRSAvailableReason is added in a deployment when its newest replica set is made available
	// ie. the number of new pods that have passed readiness checks and run for at least minReadySeconds
	// is at least the minimum available pods that need to run for the deployment.
	NewRSAvailableReason = "NewReplicaSetAvailable"

	// Available:
	// MinimumReplicasAvailable is added in a deployment when it has its minimum replicas required available.
	MinimumReplicasAvailable   = "MinimumReplicasAvailable"
	MinimumReplicasUnavailable = "MinimumReplicasUnavailable"
)

func isDeploymentReady(deployment appsv1.Deployment) bool {
	conditions := deployment.Status.Conditions
	return hasDeploymentConditionTrueStatusWithReason(conditions, appsv1.DeploymentAvailable, MinimumReplicasAvailable) &&
		hasDeploymentConditionTrueStatusWithReason(conditions, appsv1.DeploymentProgressing, NewRSAvailableReason)
}

func hasDeploymentConditionTrueStatusWithReason(conditions []appsv1.DeploymentCondition, conditionType appsv1.DeploymentConditionType, reason string) bool {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status == corev1.ConditionTrue &&
				condition.Reason == reason
		}
	}
	return false
}
