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

	if statusChanged, serverlessConfigurationMsg := updateStatus(&s.instance, eventingURL, tracingURL); statusChanged {
		s.setState(v1alpha1.StateProcessing)
		s.instance.UpdateConditionTrue(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigured,
			fmt.Sprintf(serverlessConfigurationMsg),
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
	)

	return nextState(sFnApplyResources)
}

func getTracingURL(ctx context.Context, log *zap.SugaredLogger, client client.Client, spec v1alpha1.ServerlessSpec) (string, error) {
	if spec.Tracing != nil {
		if spec.Tracing.Endpoint == "" {
			return v1alpha1.FeatureDisabled, nil
		}
		return spec.Tracing.Endpoint, nil
	}

	tracingURL, err := tracing.GetTraceCollectorURL(ctx, client)
	if err != nil {
		return "", errors.Wrap(err, "while getting trace pipeline")
	}
	if tracingURL == "" {
		return v1alpha1.FeatureDisabled, nil
	}
	return tracingURL, nil
}

func getEventingURL(spec v1alpha1.ServerlessSpec) string {
	if spec.Eventing != nil {
		if spec.Eventing.Endpoint == "" {
			return v1alpha1.FeatureDisabled
		}
		return spec.Eventing.Endpoint
	}
	return v1alpha1.DefaultEventingEndpoint
}

// returns "custom" or "no" based on args
func dependencyState(url string) string {
	switch {
	case url == "" || url == v1alpha1.FeatureDisabled:
		return "no"
	case url == v1alpha1.DefaultEventingEndpoint:
		return "default"
	default:
		return "custom"
	}
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
		{spec.CPUUtilizationPercentage, &status.CPUUtilizationPercentage, "CPU utilization: %s"},
		{spec.RequeueDuration, &status.RequeueDuration, "function requeue duration: %s"},
		{spec.BuildExecutorArgs, &status.BuildExecutorArgs, "function build executor args: %s"},
		{spec.BuildMaxSimultaneousJobs, &status.BuildMaxSimultaneousJobs, "max number of simultaneous jobs: %s"},
		{spec.HealthzLivenessTimeout, &status.HealthzLivenessTimeout, "duration of health check: %s"},
		{spec.RequestBodyLimitMb, &status.RequestBodyLimitMb, "max size of request body: %s"},
		{spec.TimeoutSec, &status.TimeoutSec, "timeout: %s"},
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
	if !meta.IsStatusConditionPresentAndEqual(instance.Status.Conditions, string(v1alpha1.ConditionTypeConfigured), metav1.ConditionTrue) {
		hasChanged = true
	}
	instance.Status = status
	return hasChanged, sb.String()
}
