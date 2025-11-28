package chart

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

func Verify(config *Config) (bool, error) {
	spec, err := config.Cache.Get(config.Ctx, config.CacheKey)
	if err != nil {
		return false, fmt.Errorf("could not render manifest from chart: %s", err.Error())
	}
	// sometimes cache is not created yet
	if len(spec.Manifest) == 0 {
		return false, nil
	}

	objs, err := parseManifest(spec.Manifest)
	if err != nil {
		return false, fmt.Errorf("could not parse chart manifest: %s", err.Error())
	}

	for i := range objs {
		u := objs[i]

		if u.GetKind() != "Deployment" {
			continue
		}

		ready, err := verifyDeployment(config, u)
		if err != nil {
			return false, fmt.Errorf("could not verify deployment %s/%s: %s", u.GetNamespace(), u.GetName(), err.Error())
		}

		if !ready {
			return false, nil
		}
	}

	return true, nil
}

func verifyDeployment(config *Config, u unstructured.Unstructured) (bool, error) {
	var deployment appsv1.Deployment
	err := config.Cluster.Client.Get(config.Ctx, types.NamespacedName{
		Name:      u.GetName(),
		Namespace: u.GetNamespace(),
	}, &deployment)
	if err != nil {
		return false, err
	}

	return isDeploymentReady(deployment), nil
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
