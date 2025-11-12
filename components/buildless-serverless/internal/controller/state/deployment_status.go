package state

import (
	"context"
	"fmt"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/metrics"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnDeploymentStatus(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	m.State.Function.Status.ObservedGeneration = m.State.Function.GetGeneration()

	clusterDeployments, err := getDeployments(ctx, m)
	if err != nil {
		return stopWithError(errors.Wrap(err, "while getting deployments"))
	}
	// reconcile again if there are multiple or no deployments
	if len(clusterDeployments.Items) != 1 {
		return stopWithError(errors.New("multiple or no deployments found"))
	}
	deployment := clusterDeployments.Items[0]
	deploymentName := deployment.GetName()
	m.State.ClusterDeployment = &deployment

	// ready deployment
	if isDeploymentReady(deployment) {
		m.Log.Info(fmt.Sprintf("deployment %s ready", deploymentName))

		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionTrue,
			serverlessv1alpha2.ConditionReasonDeploymentReady,
			fmt.Sprintf("Deployment %s is ready", deploymentName))
		metrics.PublishStateReachTime(m.State.Function, serverlessv1alpha2.ConditionRunning)

		return nextState(sFnAdjustStatus)
	}

	// unhealthy deployment
	if hasDeploymentConditionFalseStatusWithReason(deployment.Status.Conditions, appsv1.DeploymentAvailable, MinimumReplicasUnavailable) {
		m.Log.Info(fmt.Sprintf("deployment unhealthy: %q", deploymentName))

		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionUnknown,
			serverlessv1alpha2.ConditionReasonMinReplicasNotAvailable,
			fmt.Sprintf("Minimum replicas not available for deployment %s", deploymentName))

		return requeue()
	}

	// deployment not ready
	if hasDeploymentConditionTrueStatus(deployment.Status.Conditions, appsv1.DeploymentProgressing) {
		m.Log.Info(fmt.Sprintf("deployment %q not ready", deploymentName))

		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionUnknown,
			serverlessv1alpha2.ConditionReasonDeploymentWaiting,
			fmt.Sprintf("Deployment %s is not ready yet", deploymentName))

		return requeue()
	}

	// deployment failed
	m.Log.Info(fmt.Sprintf("deployment %q failed", deploymentName))

	yamlConditions, err := yaml.Marshal(deployment.Status.Conditions)
	if err != nil {
		return stopWithError(errors.Wrap(err, "while parsing deployment status"))
	}
	msg := fmt.Sprintf("Deployment %s failed with condition: \n%s", deploymentName, yamlConditions)
	m.State.Function.UpdateCondition(
		serverlessv1alpha2.ConditionRunning,
		metav1.ConditionFalse,
		serverlessv1alpha2.ConditionReasonDeploymentFailed,
		msg)

	return stop()
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

func hasDeploymentConditionFalseStatusWithReason(conditions []appsv1.DeploymentCondition, conditionType appsv1.DeploymentConditionType, reason string) bool {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status == corev1.ConditionFalse &&
				condition.Reason == reason
		}
	}
	return false
}

func hasDeploymentConditionTrueStatus(conditions []appsv1.DeploymentCondition, conditionType appsv1.DeploymentConditionType) bool {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}
