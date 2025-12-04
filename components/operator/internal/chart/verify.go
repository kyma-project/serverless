package chart

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// VerificationCompleted indicates that the verification has been completed
	VerificationCompleted = "OK"

	// DeploymentVerificationProcessing indicates that the deployment is still being processed
	DeploymentVerificationProcessing = "DeploymentProcessing"
)

type VerificationResult struct {
	Ready  bool
	Reason string
}

func Verify(config *Config) (*VerificationResult, error) {
	spec, err := config.Cache.Get(config.Ctx, config.CacheKey)
	if err != nil {
		return nil, fmt.Errorf("could not render manifest from chart: %s", err.Error())
	}
	// sometimes cache is not created yet
	if len(spec.Manifest) == 0 {
		return &VerificationResult{Ready: false}, nil
	}

	objs, err := parseManifest(spec.Manifest)
	if err != nil {
		return nil, fmt.Errorf("could not parse chart manifest: %s", err.Error())
	}

	for i := range objs {
		u := objs[i]

		if u.GetKind() != "Deployment" {
			continue
		}

		reason, err := verifyDeployment(config, u)
		if err != nil {
			return nil, fmt.Errorf("could not verify deployment %s/%s: %s", u.GetNamespace(), u.GetName(), err.Error())
		}

		if reason != VerificationCompleted {
			return &VerificationResult{Ready: false, Reason: reason}, nil
		}
	}

	return &VerificationResult{Ready: true, Reason: VerificationCompleted}, nil
}

func verifyDeployment(config *Config, u unstructured.Unstructured) (string, error) {
	var deployment appsv1.Deployment
	err := config.Cluster.Client.Get(config.Ctx, types.NamespacedName{
		Name:      u.GetName(),
		Namespace: u.GetNamespace(),
	}, &deployment)
	if err != nil {
		return "", err
	}

	if isDeploymentReady(deployment) {
		return VerificationCompleted, nil
	}

	if hasDeploymentConditionTrueStatus(deployment.Status.Conditions, appsv1.DeploymentReplicaFailure) {
		return fmt.Sprintf("deployment %s/%s has replica failure: %s", u.GetNamespace(), u.GetName(),
			getDeploymentCondition(deployment.Status.Conditions, appsv1.DeploymentReplicaFailure).Message), nil
	}

	return DeploymentVerificationProcessing, nil
}

const (
	// NewRSAvailableReason is added in a deployment when its newest replica set is made available
	// ie. the number of new pods that have passed readiness checks and run for at least minReadySeconds
	// is at least the minimum available pods that need to run for the deployment.
	NewRSAvailableReason = "NewReplicaSetAvailable"

	// MinimumReplicasAvailable is added in a deployment when it has its minimum replicas required available.
	MinimumReplicasAvailable = "MinimumReplicasAvailable"
)

func isDeploymentReady(deployment appsv1.Deployment) bool {
	conditions := deployment.Status.Conditions
	return hasDeploymentConditionTrueStatusWithReason(conditions, appsv1.DeploymentAvailable, MinimumReplicasAvailable) &&
		hasDeploymentConditionTrueStatusWithReason(conditions, appsv1.DeploymentProgressing, NewRSAvailableReason) &&
		deployment.Generation == deployment.Status.ObservedGeneration && // spec changes are observed
		deployment.Status.UnavailableReplicas == 0 // all replicas are available
}

func hasDeploymentConditionTrueStatus(conditions []appsv1.DeploymentCondition, conditionType appsv1.DeploymentConditionType) bool {
	condition := getDeploymentCondition(conditions, conditionType)
	return condition.Status == corev1.ConditionTrue
}

func hasDeploymentConditionTrueStatusWithReason(conditions []appsv1.DeploymentCondition, conditionType appsv1.DeploymentConditionType, reason string) bool {
	condition := getDeploymentCondition(conditions, conditionType)
	return condition.Status == corev1.ConditionTrue && condition.Reason == reason
}

func getDeploymentCondition(conditions []appsv1.DeploymentCondition, conditionType appsv1.DeploymentConditionType) appsv1.DeploymentCondition {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition
		}
	}
	return appsv1.DeploymentCondition{}
}
