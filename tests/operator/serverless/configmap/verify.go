package configmap

import (
	"fmt"
	"strings"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/tests/operator/utils"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FunctionConfig struct {
	Images struct {
		Nodejs20 string `yaml:"nodejs20"`
		Nodejs22 string `yaml:"nodejs22"`
		Nodejs24 string `yaml:"nodejs24"`
	} `yaml:"images"`
}

func VerifyServerlessConfigmap(testutils *utils.TestUtils, serverless *v1alpha1.Serverless) error {

	if testutils.LegacyMode {
		testutils.Logger.Info("Skipping configmap verification for legacy serverless")
		return nil
	}
	var configmap corev1.ConfigMap
	objectKey := client.ObjectKey{
		Name:      testutils.ServerlessConfigName,
		Namespace: testutils.Namespace,
	}

	err := testutils.Client.Get(testutils.Ctx, objectKey, &configmap)
	if err != nil {
		return err
	}

	return verifyConfigmap(testutils, &configmap, serverless)
}

func verifyConfigmap(testutils *utils.TestUtils, configmap *corev1.ConfigMap, serverless *v1alpha1.Serverless) error {

	raw := configmap.Data["function-config.yaml"]
	var cfg FunctionConfig
	if err := yaml.Unmarshal([]byte(raw), &cfg); err != nil {
		return err
	}

	if testutils.FipsMode {
		if !isFipsImage(cfg.Images.Nodejs20) {
			return fmt.Errorf("expected FIPS image for nodejs20, got %s", cfg.Images.Nodejs20)
		}
		if !isFipsImage(cfg.Images.Nodejs22) {
			return fmt.Errorf("expected FIPS image for nodejs22, got %s", cfg.Images.Nodejs22)
		}
		if !isFipsImage(cfg.Images.Nodejs24) {
			return fmt.Errorf("expected FIPS image for nodejs24, got %s", cfg.Images.Nodejs24)
		}
	}

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

func isFipsImage(image string) bool {
	return image != "" && strings.Contains(image, "fips") && strings.Contains(image, "restricted-prod")
}
