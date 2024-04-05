package state

import (
	"context"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

const (
	slowBuildPreset   = "slow"
	slowRuntimePreset = "XS"
	fastBuildPreset   = "fast"
	fastRuntimePreset = "L"
)

func sFnControllerConfiguration(ctx context.Context, r *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
	err := updateControllerConfigurationStatus(ctx, r, &s.instance)
	if err != nil {
		return stopWithEventualError(err)
	}

	configureControllerConfigurationFlags(s)

	s.setState(v1alpha1.StateProcessing)
	s.instance.UpdateConditionTrue(
		v1alpha1.ConditionTypeConfigured,
		v1alpha1.ConditionReasonConfigured,
		"Configuration ready",
	)

	return nextState(sFnApplyResources)
}

func updateControllerConfigurationStatus(ctx context.Context, r *reconciler, instance *v1alpha1.Serverless) error {
	spec := instance.Spec
	fields := fieldsToUpdate{
		{spec.HealthzLivenessTimeout, &instance.Status.HealthzLivenessTimeout, "Duration of health check", ""},
	}

	updateStatusFields(r.k8s, instance, fields)
	return nil
}

func configureControllerConfigurationFlags(s *systemState) {
	s.flagsBuilder.
		WithControllerConfiguration(
			s.instance.Status.CPUUtilizationPercentage,
			s.instance.Status.RequeueDuration,
			s.instance.Status.BuildExecutorArgs,
			s.instance.Status.BuildMaxSimultaneousJobs,
			s.instance.Status.HealthzLivenessTimeout,
			s.instance.Status.RequestBodyLimitMb,
			s.instance.Status.TimeoutSec,
		).
		WithDefaultPresetFlags(
			s.instance.Status.DefaultBuildJobPreset,
			s.instance.Status.DefaultRuntimePodPreset,
		)
}
