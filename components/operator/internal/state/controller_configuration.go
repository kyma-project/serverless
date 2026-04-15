package state

import (
	"context"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	slowRuntimePreset  = "XS"
	largeRuntimePreset = "L"
	defaultLogLevel    = "info"
	defaultLogFormat   = "json"
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

	return nextState(sFnConfigureNetworkPolicies)
}

func updateControllerConfigurationStatus(ctx context.Context, r *reconciler, instance *v1alpha1.Serverless) error {
	nodesLen, err := getNodesLen(ctx, r.client)
	if err != nil {
		return err
	}

	defaultRuntimePreset := slowRuntimePreset
	if nodesLen > 2 {
		defaultRuntimePreset = largeRuntimePreset
	}

	spec := instance.Spec

	fields := fieldsToUpdate{
		{spec.FunctionRequeueDuration, &instance.Status.RequeueDuration, "Function requeue duration", ""},
		{spec.HealthzLivenessTimeout, &instance.Status.HealthzLivenessTimeout, "Duration of health check", ""},
		{spec.DefaultRuntimePodPreset, &instance.Status.DefaultRuntimePodPreset, "Default runtime pod preset", defaultRuntimePreset},
		{spec.LogLevel, &instance.Status.LogLevel, "Log level", defaultLogLevel},
		{spec.LogFormat, &instance.Status.LogFormat, "Log format", defaultLogFormat},
	}

	updateStatusFields(r.k8s, instance, fields)
	return nil
}

func configureControllerConfigurationFlags(s *systemState) {
	s.flagsBuilder.
		WithControllerConfiguration(
			s.instance.Status.RequeueDuration,
			s.instance.Status.HealthzLivenessTimeout,
		).
		WithDefaultPresetFlags(
			s.instance.Status.DefaultRuntimePodPreset,
		).
		WithLogLevel(s.instance.Status.LogLevel).
		WithLogFormat(s.instance.Status.LogFormat).
		WithLogFormatRestartAnnotation(s.instance.Status.LogFormat)
}

func getNodesLen(ctx context.Context, c client.Client) (int, error) {
	nodeList := corev1.NodeList{}
	err := c.List(ctx, &nodeList)
	if err != nil {
		return 0, err
	}

	return len(nodeList.Items), nil
}
