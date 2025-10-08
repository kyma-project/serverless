package configmap

import (
	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/tests/operator/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func VerifyServerlessConfigmap(testutils *utils.TestUtils, serverless *v1alpha1.Serverless) error {
	var configmap corev1.ConfigMap
	objectKey := client.ObjectKey{
		Name:      testutils.ServerlessConfigName,
		Namespace: testutils.Namespace,
	}

	err := testutils.Client.Get(testutils.Ctx, objectKey, &configmap)
	if err != nil {
		return err
	}

	return verifyConfigmap(&configmap, serverless)
}

func verifyConfigmap(configmap *corev1.ConfigMap, serverless *v1alpha1.Serverless) error {
	//TODO:  verify if all data from the spec is reflected in the configmap

	// expectedData := map[string]interface{}{
	// 	"functionTraceCollectorEndpoint":                          serverless.Status.TracingEndpoint,
	// 	"functionPublisherProxyAddress":                           serverless.Status.EventingEndpoint,
	// 	"resourcesConfiguration.function.resources.defaultPreset": serverless.Status.DefaultRuntimePodPreset,
	// 	"functionReadyRequeueDuration":                            serverless.Status.RequeueDuration,
	// 	"healthzLivenessTimeout":                                  serverless.Status.HealthzLivenessTimeout,
	// }

	// actualStringData :=  configmap.Data["function-config.yaml"]

	return nil
}
