package state

import (
	"context"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	nodesLen, err := getNodesLen(ctx, r.client)
	if err != nil {
		return err
	}

	defaultBuildPreset := slowBuildPreset
	defaultRuntimePreset := slowRuntimePreset
	if nodesLen > 2 {
		defaultBuildPreset = fastBuildPreset
		defaultRuntimePreset = fastRuntimePreset
	}

	spec := instance.Spec
	fields := fieldsToUpdate{
		{spec.TargetCPUUtilizationPercentage, &instance.Status.CPUUtilizationPercentage, "CPU utilization", ""},
		{spec.FunctionRequeueDuration, &instance.Status.RequeueDuration, "Function requeue duration", ""},
		{spec.FunctionBuildExecutorArgs, &instance.Status.BuildExecutorArgs, "Function build executor args", ""},
		{spec.FunctionBuildMaxSimultaneousJobs, &instance.Status.BuildMaxSimultaneousJobs, "Max number of simultaneous jobs", ""},
		{spec.HealthzLivenessTimeout, &instance.Status.HealthzLivenessTimeout, "Duration of health check", ""},
		{spec.FunctionRequestBodyLimitMb, &instance.Status.RequestBodyLimitMb, "Max size of request body", ""},
		{spec.FunctionTimeoutSec, &instance.Status.TimeoutSec, "Timeout", ""},
		{spec.DefaultBuildJobPreset, &instance.Status.DefaultBuildJobPreset, "Default build job preset", defaultBuildPreset},
		{spec.DefaultRuntimePodPreset, &instance.Status.DefaultRuntimePodPreset, "Default runtime pod preset", defaultRuntimePreset},
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

func getNodesLen(ctx context.Context, c client.Client) (int, error) {
	nodeList := corev1.NodeList{}
	err := c.List(ctx, &nodeList)
	if err != nil {
		return 0, err
	}

	return len(nodeList.Items), nil
}
