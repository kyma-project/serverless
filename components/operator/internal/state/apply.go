package state

import (
	"context"
	"fmt"
	"os"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/chart"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// run serverless chart installation
func sFnApplyResources(_ context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	// set condition Installed if it does not exist
	if !s.instance.IsCondition(v1alpha1.ConditionTypeInstalled) {
		s.setState(v1alpha1.StateProcessing)
		s.instance.UpdateConditionUnknown(v1alpha1.ConditionTypeInstalled, v1alpha1.ConditionReasonInstallation,
			"Installing for configuration")
	}

	// update common labels for all rendered resources
	s.flagsBuilder.WithManagedByLabel("serverless-operator")

	// update all used images
	updateImages(s.flagsBuilder)

	// install component
	err := install(s)
	if err != nil {
		fmt.Println(err)
		r.log.Warnf("error while installing resource %s: %s",
			client.ObjectKeyFromObject(&s.instance), err.Error())
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonInstallationErr,
			err,
		)
		return stopWithEventualError(err)
	}

	// switch state verify
	return nextState(sFnVerifyResources)
}

func install(s *systemState) error {
	flags, err := s.flagsBuilder.Build()
	if err != nil {
		return err
	}

	return chart.Install(s.chartConfig, flags)
}

func updateImages(fb chart.FlagsBuilder) {
	updateImageIfOverride("IMAGE_FUNCTION_CONTROLLER", fb.WithImageFunctionBuildfulController)
	updateImageIfOverride("IMAGE_FUNCTION_BUILDLESS_CONTROLLER", fb.WithImageFunctionController)
	updateImageIfOverride("IMAGE_FUNCTION_BUILD_INIT", fb.WithImageFunctionBuildInit)
	updateImageIfOverride("IMAGE_FUNCTION_BUILDLESS_INIT", fb.WithImageFunctionInit)
	updateImageIfOverride("IMAGE_REGISTRY_INIT", fb.WithImageRegistryInit)
	updateImageIfOverride("IMAGE_FUNCTION_RUNTIME_NODEJS20", fb.WithImageFunctionRuntimeNodejs20)
	updateImageIfOverride("IMAGE_FUNCTION_RUNTIME_NODEJS22", fb.WithImageFunctionRuntimeNodejs22)
	updateImageIfOverride("IMAGE_FUNCTION_RUNTIME_PYTHON312", fb.WithImageFunctionRuntimePython312)
	updateImageIfOverride("IMAGE_KANIKO_EXECUTOR", fb.WithImageKanikoExecutor)
	updateImageIfOverride("IMAGE_REGISTRY", fb.WithImageRegistry)
}

func updateImageIfOverride(envName string, updateFunction chart.ImageReplace) {
	imageName := os.Getenv(envName)
	if imageName != "" {
		updateFunction(imageName)
	}
}
