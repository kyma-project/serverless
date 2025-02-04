package state

import (
	"context"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/controller/fsm"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	defaultRequeueTime = time.Second * 1
)

func sFnDeploymentStatus(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	deploymentName := m.State.BuiltDeployment.GetName()
	deployment := appsv1.Deployment{}
	err := m.Client.Get(ctx, client.ObjectKey{
		Namespace: m.State.BuiltDeployment.GetNamespace(),
		Name:      deploymentName,
	}, &deployment)
	if err != nil {
		return nil, &ctrl.Result{RequeueAfter: defaultRequeueTime}, errors.Wrap(err, "while getting deployments")
	}
	m.State.ClusterDeployment = &deployment

	// ready deployment
	if isDeploymentReady(deployment) {
		m.Log.Info(fmt.Sprintf("deployment %s ready", deploymentName))

		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionTrue,
			serverlessv1alpha2.ConditionReasonDeploymentReady,
			fmt.Sprintf("Deployment %s is ready", deploymentName))

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

		return requeueAfter(defaultRequeueTime)
	}

	// deployment not ready
	if hasDeploymentConditionTrueStatus(deployment.Status.Conditions, appsv1.DeploymentProgressing) {
		m.Log.Info(fmt.Sprintf("deployment %q not ready", deploymentName))

		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionUnknown,
			serverlessv1alpha2.ConditionReasonDeploymentWaiting,
			fmt.Sprintf("Deployment %s is not ready yet", deploymentName))

		return requeueAfter(defaultRequeueTime)
	}

	// deployment failed
	m.Log.Info(fmt.Sprintf("deployment %q failed", deploymentName))

	yamlConditions, err := yaml.Marshal(deployment.Status.Conditions)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while parsing deployment status")
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
