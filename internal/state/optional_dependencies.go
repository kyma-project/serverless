package state

import (
	"context"
	"fmt"
	"github.com/kyma-project/serverless-manager/internal/tracing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
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

	// update status and condition if status is not up-to-date
<<<<<<< HEAD
	if s.instance.Status.EventingEndpoint != eventingURL ||
		s.instance.Status.TracingEndpoint != tracingURL ||
		!meta.IsStatusConditionPresentAndEqual(s.instance.Status.Conditions, string(v1alpha1.ConditionTypeConfigured), metav1.ConditionTrue) {
=======
	if s.instance.Spec.Eventing != nil && s.instance.Spec.Tracing != nil && (s.instance.Status.EventingEndpoint != s.instance.Spec.Eventing.Endpoint ||
		s.instance.Status.TracingEndpoint != s.instance.Spec.Tracing.Endpoint ||
		!meta.IsStatusConditionPresentAndEqual(s.instance.Status.Conditions, string(v1alpha1.ConditionTypeConfigured), metav1.ConditionTrue)) {
>>>>>>> 9e05a5a (add test to test cr additional)

		s.instance.Status.EventingEndpoint = eventingURL
		s.instance.Status.TracingEndpoint = tracingURL

		s.setState(v1alpha1.StateProcessing)
		s.instance.UpdateConditionTrue(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigured,
			fmt.Sprintf("Configured with %s Publisher Proxy URL and %s Trace Collector URL.",
				dependencyState(s.instance.Status.EventingEndpoint),
				dependencyState(s.instance.Status.TracingEndpoint)),
		)
		return nextState(sFnUpdateStatusAndRequeue)
	}

	if (s.instance.Spec.CPUUtilizationPercentage != nil &&
		s.instance.Spec.RequeueDuration != nil &&
		s.instance.Spec.BuildExecutorArgs != nil &&
		s.instance.Spec.BuildMaxSimultaneousJobs != nil &&
		s.instance.Spec.HealthzLivenessTimeout != nil &&
		s.instance.Spec.RequestBodyLimitMb != nil &&
		s.instance.Spec.TimeoutSec != nil) &&
		(s.instance.Status.CPUUtilizationPercentage != s.instance.Spec.CPUUtilizationPercentage.AdditionalConfig ||
			s.instance.Status.RequeueDuration != s.instance.Spec.RequeueDuration.AdditionalConfig ||
			s.instance.Status.BuildExecutorArgs != s.instance.Spec.BuildExecutorArgs.AdditionalConfig ||
			s.instance.Status.BuildMaxSimultaneousJobs != s.instance.Spec.BuildMaxSimultaneousJobs.AdditionalConfig ||
			s.instance.Status.HealthzLivenessTimeout != s.instance.Spec.HealthzLivenessTimeout.AdditionalConfig ||
			s.instance.Status.RequestBodyLimitMb != s.instance.Spec.RequestBodyLimitMb.AdditionalConfig ||
			s.instance.Status.TimeoutSec != s.instance.Spec.TimeoutSec.AdditionalConfig) {

		s.instance.Status.CPUUtilizationPercentage = s.instance.Spec.CPUUtilizationPercentage.AdditionalConfig
		s.instance.Status.RequeueDuration = s.instance.Spec.RequeueDuration.AdditionalConfig
		s.instance.Status.BuildExecutorArgs = s.instance.Spec.BuildExecutorArgs.AdditionalConfig
		s.instance.Status.BuildMaxSimultaneousJobs = s.instance.Spec.BuildMaxSimultaneousJobs.AdditionalConfig
		s.instance.Status.HealthzLivenessTimeout = s.instance.Spec.HealthzLivenessTimeout.AdditionalConfig
		s.instance.Status.RequestBodyLimitMb = s.instance.Spec.RequestBodyLimitMb.AdditionalConfig
		s.instance.Status.TimeoutSec = s.instance.Spec.TimeoutSec.AdditionalConfig

		s.setState(v1alpha1.StateProcessing)
		s.instance.UpdateConditionTrue(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigured,
			fmt.Sprintf("Configured with %s CPU utilization, %s function requeue duration, %s function build executor args, %s max number of simultaneous jobs, %s duration of health check, %s max size of request body and %s timeout.",
				dependencyState(s.instance.Status.CPUUtilizationPercentage, v1alpha1.DefaultCPUUtilizationPercentage),
				dependencyState(s.instance.Status.RequeueDuration, v1alpha1.DefaultRequeueDuration),
				dependencyState(s.instance.Status.BuildExecutorArgs, v1alpha1.DefaultBuildExecutorArgs),
				dependencyState(s.instance.Status.BuildMaxSimultaneousJobs, v1alpha1.DefaultBuildMaxSimultaneousJobs),
				dependencyState(s.instance.Status.HealthzLivenessTimeout, v1alpha1.DefaultHealthzLivenessTimeout),
				dependencyState(s.instance.Status.RequestBodyLimitMb, v1alpha1.DefaultRequestBodyLimitMb),
				dependencyState(s.instance.Status.TimeoutSec, v1alpha1.DefaultTimeoutSec)),
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
