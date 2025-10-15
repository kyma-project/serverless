package serverless

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/tests/operator/serverless/configmap"
	"github.com/kyma-project/serverless/tests/operator/serverless/deployment"
	"github.com/kyma-project/serverless/tests/operator/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func VerifyDeletion(utils *utils.TestUtils) error {
	err := Verify(utils)
	if err == nil {
		return fmt.Errorf("serverless '%s' still exists", utils.ServerlessName)
	}
	if !errors.IsNotFound(err) {
		return err
	}

	return nil
}

func Verify(utils *utils.TestUtils) error {
	serverless, err := getServerless(utils, utils.ServerlessName)
	if err != nil {
		return err
	}

	if err := verifyState(utils, &serverless); err != nil {
		return err
	}

	if err := verifyStatus(&serverless, utils.LegacyMode); err != nil {
		return err
	}

	if utils.LegacyMode {
		return deployment.VerifyCtrlMngrEnvs(utils, &serverless)
	}

	return configmap.VerifyServerlessConfigmap(utils, &serverless)
}

func VerifyConfig(utils *utils.TestUtils) error {
	configMap := &corev1.ConfigMap{}
	objectKey := client.ObjectKey{
		Name:      utils.ServerlessConfigMapName,
		Namespace: utils.Namespace,
	}
	err := utils.Client.Get(utils.Ctx, objectKey, configMap)
	return err
}

func VerifyStuck(utils *utils.TestUtils) error {
	serverless, err := getServerless(utils, utils.SecondServerlessName)
	if err != nil {
		return err
	}

	if err := verifyStateStuck(utils, &serverless); err != nil {
		return err
	}

	return nil
}

func VerifyDeletionStuck(utils *utils.TestUtils) error {
	serverless, err := getServerless(utils, utils.ServerlessName)
	if err != nil {
		return err
	}

	return verifyDeletionStuck(&serverless)
}

func getServerless(utils *utils.TestUtils, name string) (v1alpha1.Serverless, error) {
	var serverless v1alpha1.Serverless
	objectKey := client.ObjectKey{
		Name:      name,
		Namespace: utils.Namespace,
	}

	if err := utils.Client.Get(utils.Ctx, objectKey, &serverless); err != nil {
		return v1alpha1.Serverless{}, err
	}

	return serverless, nil
}

// check if all data from the spec is reflected in the status
func verifyStatus(serverless *v1alpha1.Serverless, legacy bool) error {
	status := serverless.Status
	spec := serverless.Spec

	if err := isSpecValueReflectedInStatus(spec.FunctionRequeueDuration, status.RequeueDuration); err != nil {
		return err
	}

	if err := isSpecValueReflectedInStatus(spec.HealthzLivenessTimeout, status.HealthzLivenessTimeout); err != nil {
		return err
	}

	if err := isSpecValueReflectedInStatus(spec.DefaultRuntimePodPreset, status.DefaultRuntimePodPreset); err != nil {
		return err
	}

	if err := isSpecValueReflectedInStatus(spec.LogLevel, status.LogLevel); err != nil {
		return err
	}

	if err := isSpecValueReflectedInStatus(spec.LogFormat, status.LogFormat); err != nil {
		return err
	}

	if spec.Eventing != nil {
		if err := isSpecValueReflectedInStatus(spec.Eventing.Endpoint, status.EventingEndpoint); err != nil {
			return err
		}
	}

	if spec.Tracing != nil {
		if err := isSpecValueReflectedInStatus(spec.Tracing.Endpoint, status.TracingEndpoint); err != nil {
			return err
		}
	}

	if err := isSpecBooleanValueReflectedInStatus(spec.EnableNetworkPolicies, status.NetworkPoliciesEnabled); err != nil {
		return err
	}

	if legacy {

		if err := isSpecValueReflectedInStatus(spec.TargetCPUUtilizationPercentage, status.CPUUtilizationPercentage); err != nil {
			return err
		}

		if err := isSpecValueReflectedInStatus(spec.FunctionBuildExecutorArgs, status.BuildExecutorArgs); err != nil {
			return err
		}

		if err := isSpecValueReflectedInStatus(spec.FunctionBuildMaxSimultaneousJobs, status.BuildMaxSimultaneousJobs); err != nil {
			return err
		}

		if err := isSpecValueReflectedInStatus(spec.DefaultBuildJobPreset, status.DefaultBuildJobPreset); err != nil {
			return err
		}
	}

	return nil
}

func isSpecValueReflectedInStatus(specValue string, statusValue string) error {
	if specValue == "" {
		// value is not set in the spec, so value in the status may be empty or defauled
		return nil
	}

	if specValue != statusValue {
		return fmt.Errorf("value '%s' not found in status", specValue)
	}

	return nil
}

func isSpecBooleanValueReflectedInStatus(specFlag bool, statusValue string) error {
	expectedStatusValue := "False"
	if specFlag {
		expectedStatusValue = "True"
	}

	if expectedStatusValue != statusValue {
		return fmt.Errorf("expected value '%s' not found in status", expectedStatusValue)
	}

	return nil
}

func verifyState(utils *utils.TestUtils, serverless *v1alpha1.Serverless) error {
	if serverless.Status.State != v1alpha1.StateReady {
		return fmt.Errorf("serverless '%s' in '%s' state", utils.ServerlessName, serverless.Status.State)
	}
	return nil
}

func verifyStateStuck(utils *utils.TestUtils, serverless *v1alpha1.Serverless) error {
	for _, condition := range serverless.Status.Conditions {
		if condition.Type == string(v1alpha1.ConditionTypeConfigured) {
			if condition.Reason == string(v1alpha1.ConditionReasonServerlessDuplicated) &&
				condition.Status == metav1.ConditionFalse &&
				condition.Message == fmt.Sprintf("only one instance of Serverless is allowed (current served instance: %s/%s) - this Serverless CR is redundant - remove it to fix the problem", utils.Namespace, utils.ServerlessName) {
				return nil
			}
			return fmt.Errorf("ConditionConfigured is not in expected state: %v", condition)
		}
	}
	return fmt.Errorf("ConditionConfigured not found")
}

func verifyDeletionStuck(serverless *v1alpha1.Serverless) error {
	for _, condition := range serverless.Status.Conditions {
		if condition.Type == string(v1alpha1.ConditionTypeDeleted) {
			if condition.Reason == string(v1alpha1.ConditionReasonDeletionErr) &&
				condition.Status == metav1.ConditionFalse &&
				condition.Message == "found 1 items with VersionKind serverless.kyma-project.io/v1alpha2" {
				return nil
			}
			return fmt.Errorf("ConditionDeleted is not in expected state: %v", condition)
		}
	}
	return fmt.Errorf("ConditionDeleted not found")
}
