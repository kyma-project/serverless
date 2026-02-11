package state

import (
	"context"
	"fmt"
	"os"

	"github.com/kyma-project/manager-toolkit/installation/base/resource"
	"github.com/kyma-project/manager-toolkit/installation/chart"
	"github.com/kyma-project/manager-toolkit/installation/chart/action"
	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/flags"
	"github.com/kyma-project/serverless/components/operator/internal/legacy"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	fipsVariantImageEnvKeySuffix = "_FIPS"
)

// run serverless chart installation
func sFnApplyResources(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	// set condition Installed if it does not exist
	if !s.instance.IsCondition(v1alpha1.ConditionTypeInstalled) {
		s.setState(v1alpha1.StateProcessing)
		s.instance.UpdateConditionUnknown(v1alpha1.ConditionTypeInstalled, v1alpha1.ConditionReasonInstallation,
			"Installing for configuration")
	}

	// update common labels for all rendered resources
	s.flagsBuilder.WithManagedByLabel("serverless-operator")

	// update all used images
	fipsModeEnabled := r.cfg.kymaFipsEnabled
	updateImages(fipsModeEnabled, s.flagsBuilder)

	// install component
	err := install(ctx, r, s)
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

func install(ctx context.Context, r *reconciler, s *systemState) error {
	flags, err := s.flagsBuilder.Build()
	if err != nil {
		return err
	}

	return chart.Install(s.chartConfig, &chart.InstallOpts{
		CustomFlags: flags,
		PreActions: []action.PreApply{
			// TODO: remove this callback after deleting legacy serverless
			action.PreApplyWithPredicate(
				adjustPVCPreApplyAction(ctx, r.client),
				resource.HasKind("PersistentVolumeClaim"),
			),
		},
	})
}

func updateImages(fipsModeEnabled bool, fb *flags.Builder) {
	updateImageIfOverride(fipsModeEnabled, "IMAGE_FUNCTION_CONTROLLER", fb.WithImageFunctionBuildfulController)
	updateImageIfOverride(fipsModeEnabled, "IMAGE_FUNCTION_BUILDLESS_CONTROLLER", fb.WithImageFunctionController)
	updateImageIfOverride(fipsModeEnabled, "IMAGE_FUNCTION_BUILD_INIT", fb.WithImageFunctionBuildInit)
	updateImageIfOverride(fipsModeEnabled, "IMAGE_FUNCTION_BUILDLESS_INIT", fb.WithImageFunctionInit)
	updateImageIfOverride(fipsModeEnabled, "IMAGE_REGISTRY_INIT", fb.WithImageRegistryInit)
	updateImageIfOverride(fipsModeEnabled, "IMAGE_FUNCTION_RUNTIME_NODEJS20", fb.WithImageFunctionRuntimeNodejs20)
	updateImageIfOverride(fipsModeEnabled, "IMAGE_FUNCTION_RUNTIME_NODEJS22", fb.WithImageFunctionRuntimeNodejs22)
	updateImageIfOverride(fipsModeEnabled, "IMAGE_FUNCTION_RUNTIME_NODEJS24", fb.WithImageFunctionRuntimeNodejs24)
	updateImageIfOverride(fipsModeEnabled, "IMAGE_FUNCTION_RUNTIME_PYTHON312", fb.WithImageFunctionRuntimePython312)
	updateImageIfOverride(fipsModeEnabled, "IMAGE_KANIKO_EXECUTOR", fb.WithImageKanikoExecutor)
	updateImageIfOverride(fipsModeEnabled, "IMAGE_REGISTRY", fb.WithImageRegistry)
}

func updateImageIfOverride(fipsModeEnabled bool, envName string, updateFunction flags.ImageReplace) {
	imageName := os.Getenv(envName)
	if fipsModeEnabled {
		fipsImageName := getFipsVariantImageEnv(envName)
		if fipsImageName != "" {
			imageName = fipsImageName
		}
	}

	if imageName != "" {
		updateFunction(imageName)
	}
}

func getFipsVariantImageEnv(envName string) string {
	return os.Getenv(envName + fipsVariantImageEnvKeySuffix)
}

func adjustPVCPreApplyAction(ctx context.Context, c client.Client) action.PreApply {
	return func(u *unstructured.Unstructured) error {
		adjusted, err := legacy.AdjustDockerRegToClusterPVCSize(ctx, c, *u)
		*u = adjusted
		return err
	}
}
