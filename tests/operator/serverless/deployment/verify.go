package deployment

import (
	"fmt"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/tests/operator/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func VerifyCtrlMngrEnvs(testutils *utils.TestUtils, serverless *v1alpha1.Serverless) error {
	var deploy appsv1.Deployment
	objectKey := client.ObjectKey{
		Name:      testutils.ServerlessCtrlDeployName,
		Namespace: testutils.Namespace,
	}

	err := testutils.Client.Get(testutils.Ctx, objectKey, &deploy)
	if err != nil {
		return err
	}

	if testutils.FipsMode {
		return verifyFipsModeDeployEnvs(&deploy, serverless)
	}

	if testutils.LegacyMode {
		return verifyLegacyDeployEnvs(&deploy, serverless)
	}

	return nil
}

func VerifyFipsModeCtrlMngrEnvs(testutils *utils.TestUtils, serverless *v1alpha1.Serverless) error {
	var deploy appsv1.Deployment
	objectKey := client.ObjectKey{
		Name:      testutils.ServerlessCtrlDeployName,
		Namespace: testutils.Namespace,
	}

	err := testutils.Client.Get(testutils.Ctx, objectKey, &deploy)
	if err != nil {
		return err
	}

	return verifyFipsModeDeployEnvs(&deploy, serverless)
}

func VerifyCtrlMngrAnnotations(testutils *utils.TestUtils) error {
	if testutils.LegacyMode {
		// in legacy mode annotations are not applied, so skip the check
		return nil
	}
	var deploy appsv1.Deployment
	objectKey := client.ObjectKey{
		Name:      testutils.ServerlessCtrlDeployName,
		Namespace: testutils.Namespace,
	}

	err := testutils.Client.Get(testutils.Ctx, objectKey, &deploy)
	if err != nil {
		return err
	}

	return verifyPodTemplateAnnotations(&deploy.Spec.Template)
}

func verifyPodTemplateAnnotations(podTemplate *corev1.PodTemplateSpec) error {
	expectedAnnotations := map[string]string{
		"kubectl.kubernetes.io/default-container":    "manager",
		"sidecar.istio.io/inject":                    "false",
		"serverless.kyma-project.io/log-format":      "json",
		"rt-cfg.kyma-project.io/add-img-pull-secret": "true",
		"rt-cfg.kyma-project.io/alter-img-registry":  "true",
	}
	for key, value := range expectedAnnotations {
		if podTemplate.ObjectMeta.Annotations[key] != value {
			return fmt.Errorf("annotation '%s' with value '%s' not found in pod template", key, value)
		}
	}

	return nil
}

func verifyFipsModeDeployEnvs(deploy *appsv1.Deployment, serverless *v1alpha1.Serverless) error {
	expectedEnvs := []corev1.EnvVar{
		{
			Name:  "APP_KYMA_FIPS_MODE_ENABLED",
			Value: "true",
		},
		{
			Name:  "GODEBUG",
			Value: "fips140=only,tlsmlkem=0",
		},
	}
	for _, expectedEnv := range expectedEnvs {
		if !isEnvReflected(expectedEnv, &deploy.Spec.Template.Spec.Containers[0]) {
			return fmt.Errorf("env '%s' with value '%s' not found in deployment", expectedEnv.Name, expectedEnv.Value)
		}
	}

	return nil
}

func verifyLegacyDeployEnvs(deploy *appsv1.Deployment, serverless *v1alpha1.Serverless) error {
	expectedEnvs := []corev1.EnvVar{
		{
			Name:  "APP_FUNCTION_TRACE_COLLECTOR_ENDPOINT",
			Value: serverless.Status.TracingEndpoint,
		},
		{
			Name:  "APP_FUNCTION_PUBLISHER_PROXY_ADDRESS",
			Value: serverless.Status.EventingEndpoint,
		},
		{
			Name:  "APP_FUNCTION_TARGET_CPU_UTILIZATION_PERCENTAGE",
			Value: serverless.Status.CPUUtilizationPercentage,
		},
		{
			Name:  "APP_FUNCTION_REQUEUE_DURATION",
			Value: serverless.Status.RequeueDuration,
		},
		{
			Name:  "APP_FUNCTION_BUILD_EXECUTOR_ARGS",
			Value: serverless.Status.BuildExecutorArgs,
		},
		{
			Name:  "APP_FUNCTION_BUILD_MAX_SIMULTANEOUS_JOBS",
			Value: serverless.Status.BuildMaxSimultaneousJobs,
		},
		{
			Name:  "APP_HEALTHZ_LIVENESS_TIMEOUT",
			Value: serverless.Status.HealthzLivenessTimeout,
		},
	}
	for _, expectedEnv := range expectedEnvs {
		if !isEnvReflected(expectedEnv, &deploy.Spec.Template.Spec.Containers[0]) {
			return fmt.Errorf("env '%s' with value '%s' not found in deployment", expectedEnv.Name, expectedEnv.Value)
		}
	}

	return nil
}

func isEnvReflected(expected corev1.EnvVar, in *corev1.Container) bool {
	if expected.Value == "" {
		// return true if value is not overrided
		return true
	}

	for _, env := range in.Env {
		if env.Name == expected.Name {
			// return true if value is the same
			return env.Value == expected.Value
		}
	}

	return false
}
