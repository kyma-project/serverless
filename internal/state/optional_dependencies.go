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
	if s.instance.Status.EventingEndpoint != eventingURL ||
		s.instance.Status.TracingEndpoint != tracingURL ||
		!meta.IsStatusConditionPresentAndEqual(s.instance.Status.Conditions, string(v1alpha1.ConditionTypeConfigured), metav1.ConditionTrue) {

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

	if s.instance.Status.CPUUtilizationPercentage != s.instance.Spec.CPUUtilizationPercentage.AdditionalConfig ||
		s.instance.Status.RequeueDuration != s.instance.Spec.RequeueDuration.AdditionalConfig ||
		s.instance.Status.BuildExecutorArgs != s.instance.Spec.BuildExecutorArgs.AdditionalConfig ||
		s.instance.Status.BuildMaxSimultaneousJobs != s.instance.Spec.BuildMaxSimultaneousJobs.AdditionalConfig ||
		s.instance.Status.HealthzLivenessTimeout != s.instance.Spec.HealthzLivenessTimeout.AdditionalConfig ||
		s.instance.Status.RequestBodyLimitMb != s.instance.Spec.RequestBodyLimitMb.AdditionalConfig ||
		s.instance.Status.TimeoutSec != s.instance.Spec.TimeoutSec.AdditionalConfig {

		s.instance.Status.CPUUtilizationPercentage = s.instance.Spec.CPUUtilizationPercentage.AdditionalConfig
		s.instance.Status.RequeueDuration = s.instance.Spec.RequeueDuration.AdditionalConfig
		s.instance.Status.BuildExecutorArgs = s.instance.Spec.BuildExecutorArgs.AdditionalConfig
		s.instance.Status.BuildMaxSimultaneousJobs = s.instance.Spec.BuildMaxSimultaneousJobs.AdditionalConfig
		s.instance.Status.HealthzLivenessTimeout = s.instance.Spec.HealthzLivenessTimeout.AdditionalConfig
		s.instance.Status.RequestBodyLimitMb = s.instance.Spec.RequestBodyLimitMb.AdditionalConfig
		s.instance.Status.TimeoutSec = s.instance.Spec.TimeoutSec.AdditionalConfig
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
