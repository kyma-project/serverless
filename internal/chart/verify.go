package chart

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var (
	fromUnstructured = runtime.DefaultUnstructuredConverter.FromUnstructured
)

func Verify(config *Config) (bool, error) {
	manifest, err := getManifest(config)
	if err != nil {
		return false, fmt.Errorf("could not render manifest from chart: %s", err.Error())
	}

	objs, err := parseManifest(manifest)
	if err != nil {
		return false, fmt.Errorf("could not parse chart manifest: %s", err.Error())
	}

	for i := range objs {
		u := objs[i]

		var verifyFunc verifyFunc
		switch u.GetKind() {
		case "Deployment":
			verifyFunc = verifyDeployment
		case "DaemonSet":
			// TODO: right now we don't support internal docker registry
		default:
			continue
		}

		ready, err := verifyFunc(config, u)
		if err != nil {
			return false, fmt.Errorf("could not verify object %s/%s: %s", u.GetNamespace(), u.GetName(), err.Error())
		}

		if !ready {
			return false, nil
		}
	}

	return true, nil
}

type verifyFunc func(*Config, unstructured.Unstructured) (bool, error)

func verifyDeployment(config *Config, u unstructured.Unstructured) (bool, error) {
	var deployment appsv1.Deployment
	err := config.Client.Get(config.Ctx, types.NamespacedName{
		Name:      u.GetName(),
		Namespace: u.GetNamespace(),
	}, &deployment)
	if err != nil {
		return false, err
	}

	for _, cond := range deployment.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable && cond.Status != v1.ConditionTrue {
			return false, nil
		}
	}

	return true, nil
}
