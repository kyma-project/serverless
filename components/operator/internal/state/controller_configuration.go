package state

import (
	"context"

	"go.uber.org/zap"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	slowBuildPreset         = "slow"
	slowRuntimePreset       = "XS"
	normalBuildPreset       = "normal"
	largeRuntimePreset      = "L"
	defaultLogLevel         = "info"
	defaultLogFormat        = "json"
	buildlessChartPath      = "/buildless-module-chart"
	oldServerlessChartPath  = "/module-chart"
	buildlessModeAnnotation = "serverless.kyma-project.io/buildless-mode"
	buildlessModeEnabled    = "enabled"
	buildlessModeDisabled   = "disabled"
)

func sFnControllerConfiguration(ctx context.Context, r *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
	err := updateControllerConfigurationStatus(ctx, r, &s.instance)
	if err != nil {
		return stopWithEventualError(err)
	}

	configureControllerConfigurationFlags(s)

	//TODO: remove when buildless is enabled by default
	configureChartPathAndFlag(s, r.log)

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

	defaultBuildPreset := slowBuildPreset
	defaultRuntimePreset := slowRuntimePreset
	if nodesLen > 2 {
		defaultBuildPreset = normalBuildPreset
		defaultRuntimePreset = largeRuntimePreset
	}

	spec := instance.Spec
	fields := fieldsToUpdate{
		{spec.TargetCPUUtilizationPercentage, &instance.Status.CPUUtilizationPercentage, "CPU utilization", ""},
		{spec.FunctionRequeueDuration, &instance.Status.RequeueDuration, "Function requeue duration", ""},
		{spec.FunctionBuildExecutorArgs, &instance.Status.BuildExecutorArgs, "Function build executor args", ""},
		{spec.FunctionBuildMaxSimultaneousJobs, &instance.Status.BuildMaxSimultaneousJobs, "Max number of simultaneous jobs", ""},
		{spec.HealthzLivenessTimeout, &instance.Status.HealthzLivenessTimeout, "Duration of health check", ""},
		{spec.DefaultBuildJobPreset, &instance.Status.DefaultBuildJobPreset, "Default build job preset", defaultBuildPreset},
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
			s.instance.Status.CPUUtilizationPercentage,
			s.instance.Status.RequeueDuration,
			s.instance.Status.BuildExecutorArgs,
			s.instance.Status.BuildMaxSimultaneousJobs,
			s.instance.Status.HealthzLivenessTimeout,
		).
		WithDefaultPresetFlags(
			s.instance.Status.DefaultBuildJobPreset,
			s.instance.Status.DefaultRuntimePodPreset,
		).
		WithLogLevel(s.instance.Status.LogLevel).
		WithLogFormat(s.instance.Status.LogFormat)
}

func getNodesLen(ctx context.Context, c client.Client) (int, error) {
	nodeList := corev1.NodeList{}
	err := c.List(ctx, &nodeList)
	if err != nil {
		return 0, err
	}

	return len(nodeList.Items), nil
}

// TODO: remove these methods when buildless is enabled by default
func configureChartPathAndFlag(s *systemState, log *zap.SugaredLogger) {
	configureChartPath(s, log)
	if s.chartConfig == nil {
		return
	}
	s.flagsBuilder.WithChartPath(s.chartConfig.Release.ChartPath)
}

func configureChartPath(s *systemState, log *zap.SugaredLogger) {
	val, exists := s.instance.Annotations[buildlessModeAnnotation]
	if !exists {
		// we use default value from environment variable if annotation is not set
		return
	}
	if val == buildlessModeDisabled {
		log.Info("Chart path is set to old serverless module chart")
		s.chartConfig.Release.ChartPath = oldServerlessChartPath
	}
	log.Infof("Using chart path: %s", s.chartConfig.Release.ChartPath)
	// we use default value from environment variable if annotation has unexpected value
}
