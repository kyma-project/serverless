package state

import (
	"context"
	"fmt"
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/kyma-project/serverless-manager/internal/tracing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// enable or disable serverless optional dependencies based on the Serverless Spec and installed module on the cluster
func sFnOptionalDependencies(ctx context.Context, r *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
	// TODO: add functionality of auto-detecting these dependencies by checking Eventing CRs if user does not override these values.
	// checking these URLs manually is not possible because of lack of istio-sidecar in the serverless-operator

	tracingURL, err := getTracingURL(ctx, r.log, r.client, s.instance.Spec)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while fetching tracing URL")
	}
	eventingURL := getEventingURL(s.instance.Spec)

	if statusChanged, svlsCfgMsg := updateStatus(&s.instance, eventingURL, tracingURL); statusChanged {
		s.setState(v1alpha1.StateProcessing)
		s.instance.UpdateConditionTrue(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigured,
			fmt.Sprintf(svlsCfgMsg),
		)
		return nextState(sFnUpdateStatusAndRequeue)
	}

	if configurationStatusIsNotReady(s.instance.Status.Conditions) {
		s.setState(v1alpha1.StateProcessing)
		s.instance.UpdateConditionTrue(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigured,
			"Configuration ready",
		)
		return nextState(sFnUpdateStatusAndRequeue)
	}

	s.chartConfig.Release.Flags = chart.AppendContainersFlags(
		s.chartConfig.Release.Flags,
		s.instance.Status.EventingEndpoint,
		s.instance.Status.TracingEndpoint,
		s.instance.Status.CPUUtilizationPercentage,
		s.instance.Status.RequeueDuration,
		s.instance.Status.BuildExecutorArgs,
		s.instance.Status.BuildMaxSimultaneousJobs,
		s.instance.Status.HealthzLivenessTimeout,
		s.instance.Status.RequestBodyLimitMb,
		s.instance.Status.TimeoutSec,
		s.instance.Status.DefaultBuildJobPreset,
		s.instance.Status.DefaultRuntimePodPreset,
	)

	return nextState(sFnApplyResources)
}

func configurationStatusIsNotReady(conditions []metav1.Condition) bool {
	if !meta.IsStatusConditionPresentAndEqual(conditions, string(v1alpha1.ConditionTypeConfigured), metav1.ConditionTrue) {
		return true
	}
	return false
}

func getTracingURL(ctx context.Context, log *zap.SugaredLogger, client client.Client, spec v1alpha1.ServerlessSpec) (string, error) {
	if spec.Tracing != nil {
		return spec.Tracing.Endpoint, nil
	}

	tracingURL, err := tracing.GetTraceCollectorURL(ctx, client)
	if err != nil {
		return "", errors.Wrap(err, "while getting trace pipeline")
	}
	return tracingURL, nil
}

func getEventingURL(spec v1alpha1.ServerlessSpec) string {
	if spec.Eventing != nil {
		return spec.Eventing.Endpoint
	}
	return v1alpha1.DefaultEventingEndpoint
}

func updateStatus(instance *v1alpha1.Serverless, eventingURL, tracingURL string) (bool, string) {
	spec := instance.Spec
	status := instance.Status

	hasChanged := false

	fields := []struct {
		specField   string
		statusField *string
		cfgMsg      string
	}{
		{spec.TargetCPUUtilizationPercentage, &status.CPUUtilizationPercentage, "CPU utilization: %s"},
		{spec.FunctionRequeueDuration, &status.RequeueDuration, "function requeue duration: %s"},
		{spec.FunctionBuildExecutorArgs, &status.BuildExecutorArgs, "function build executor args: %s"},
		{spec.FunctionBuildMaxSimultaneousJobs, &status.BuildMaxSimultaneousJobs, "max number of simultaneous jobs: %s"},
		{spec.HealthzLivenessTimeout, &status.HealthzLivenessTimeout, "duration of health check: %s"},
		{spec.FunctionRequestBodyLimitMb, &status.RequestBodyLimitMb, "max size of request body: %s"},
		{spec.FunctionTimeoutSec, &status.TimeoutSec, "timeout: %s"},
		{spec.DefaultBuildJobPreset, &status.DefaultBuildJobPreset, "default build job preset: %s"},
		{spec.DefaultRuntimePodPreset, &status.DefaultRuntimePodPreset, "default runtime pod preset: %s"},
		{eventingURL, &status.EventingEndpoint, "eventing endpoint: %s"},
		{tracingURL, &status.TracingEndpoint, "tracing endpoint: %s"},
	}

	sb := strings.Builder{}
	sb.WriteString("Serverless configuration changes: ")
	separator := false

	for _, field := range fields {
		if field.specField != *field.statusField {
			if separator {
				sb.WriteString(", ")
			}
			*field.statusField = field.specField
			sb.WriteString(fmt.Sprintf(field.cfgMsg, field.specField))
			separator = true
			hasChanged = true
		}
	}
	if !hasChanged {
		sb.WriteString("no changes")
	}
	instance.Status = status

	return hasChanged, sb.String()
}
