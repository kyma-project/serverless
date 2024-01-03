package chart

import (
	"fmt"
	"runtime"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
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

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	config.Log.Warnf(printMemUsage(&m, "during verify"))

	config.Log.Infof("jestem tutaj v1")
	for i := range objs {
		u := objs[i]
		config.Log.Infof("lece w forze %s", u.GetKind())
		var verifyFunc verifyFunc
		switch u.GetKind() {
		case "Deployment":
			config.Log.Infof("jestem tutaj %s", u.GetName())
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

func printMemUsage(m *runtime.MemStats, message string) string {
	return fmt.Sprintf("{\"message\": \"%s\", \"alloc\": \"%v MiB\", \"ttotalAlloc\": \"%v MiB\", \"sys\": \"%v MiB\", \"numGC\": \"%v\"}",
		message, m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
}

type verifyFunc func(*Config, unstructured.Unstructured) (bool, error)

func verifyDeployment(config *Config, u unstructured.Unstructured) (bool, error) {
	var deployment appsv1.Deployment
	err := config.Cluster.Client.Get(config.Ctx, types.NamespacedName{
		Name:      u.GetName(),
		Namespace: u.GetNamespace(),
	}, &deployment)
	if err != nil {
		return false, err
	}

	config.Log.Infof("jestem tutaj v2 %d", len(deployment.Status.Conditions))
	for _, cond := range deployment.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable && cond.Status == v1.ConditionTrue {
			return true, nil
		}
	}

	return false, nil
}
