package serverless

import (
	"fmt"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/tests/operator/serverless/configmap"
	"github.com/kyma-project/serverless/tests/operator/serverless/deployment"
	"github.com/kyma-project/serverless/tests/operator/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func VerifyDeletion(utils *utils.TestUtils) error {
	err := Verify(utils)
	if !errors.IsNotFound(err) {
		return err
	}

	return nil
}

func Verify(utils *utils.TestUtils) error {
	var serverless v1alpha1.Serverless
	objectKey := client.ObjectKey{
		Name:      utils.ServerlessName,
		Namespace: utils.Namespace,
	}

	if err := utils.Client.Get(utils.Ctx, objectKey, &serverless); err != nil {
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

		if err := isSpecValueReflectedInStatus("internal", status.DockerRegistry); err != nil {
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
