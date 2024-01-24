package state

import (
	"context"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func sFnControllerConfiguration(_ context.Context, r *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
	updateControllerConfigurationStatus(r.k8s, &s.instance)
	configureControllerConfigurationFlags(s)

	s.setState(v1alpha1.StateProcessing)
	s.instance.UpdateConditionTrue(
		v1alpha1.ConditionTypeConfigured,
		v1alpha1.ConditionReasonConfigured,
		"Configuration ready",
	)

	return nextState(sFnApplyResources)
}

func updateControllerConfigurationStatus(eventRecorder record.EventRecorder, instance *v1alpha1.Serverless) {
	spec := instance.Spec

	fields := fieldsToUpdate{
		{spec.TargetCPUUtilizationPercentage, &instance.Status.CPUUtilizationPercentage, "CPU utilization"},
		{spec.FunctionRequeueDuration, &instance.Status.RequeueDuration, "Function requeue duration"},
		{spec.FunctionBuildExecutorArgs, &instance.Status.BuildExecutorArgs, "Function build executor args"},
		{spec.FunctionBuildMaxSimultaneousJobs, &instance.Status.BuildMaxSimultaneousJobs, "Max number of simultaneous jobs"},
		{spec.HealthzLivenessTimeout, &instance.Status.HealthzLivenessTimeout, "Duration of health check"},
		{spec.FunctionRequestBodyLimitMb, &instance.Status.RequestBodyLimitMb, "Max size of request body"},
		{spec.FunctionTimeoutSec, &instance.Status.TimeoutSec, "Timeout"},
		{spec.DefaultBuildJobPreset, &instance.Status.DefaultBuildJobPreset, "Default build job preset"},
		{spec.DefaultRuntimePodPreset, &instance.Status.DefaultRuntimePodPreset, "Default runtime pod preset"},
	}

	updateStatusFields(eventRecorder, instance, fields)
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
