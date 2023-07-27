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

const Disabled = "disabled"

// enable or disable serverless optional dependencies based on the Serverless Spec and installed module on the cluster
func sFnOptionalDependencies(ctx context.Context, r *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
	// TODO: add functionality of auto-detecting these dependencies by checking Eventing CRs if user does not override these values.
	// checking these URLs manually is not possible because of lack of istio-sidecar in the serverless-operator

	tracingURL, err := getTracingUrl(ctx, r.log, r.client, s.instance.Spec)
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
				dependencyState(s.instance.Status.EventingEndpoint, v1alpha1.DefaultPublisherProxyURL),
				dependencyState(s.instance.Status.TracingEndpoint, v1alpha1.DefaultTraceCollectorURL)),
		)
		return nextState(sFnUpdateStatusAndRequeue)
	}

	s.chartConfig.Release.Flags = chart.AppendContainersFlags(
		s.chartConfig.Release.Flags,
		s.instance.Status.EventingEndpoint,
		s.instance.Status.TracingEndpoint,
	)

	return nextState(sFnApplyResources)
}

func getTracingUrl(ctx context.Context, log *zap.SugaredLogger, client client.Client, spec v1alpha1.ServerlessSpec) (string, error) {
	if spec.Tracing != nil {
		return spec.Tracing.Endpoint, nil
	}

	tracePipelines, err := tracing.GetTracePipeline(ctx, client)
	if err != nil {
		return "", errors.Wrap(err, "while getting trace pipeline")
	}

	tracePipelinesLen := len(tracePipelines.Items)
	if tracePipelinesLen != 1 {
		log.Warnf("Cluster has %d TracePipelines, it should have only one", tracePipelinesLen)
	}
	if tracePipelinesLen == 0 {
		return "", nil
	}
	otlp := tracePipelines.Items[0].Spec.Output.Otlp
	if otlp == nil {
		log.Warn("Trace pipeline is misconfigured, couldn't find otlp")
		return "", nil
	}

	return otlp.Endpoint.Value, nil
}

func getEventingURL(spec v1alpha1.ServerlessSpec) string {
	if spec.Eventing != nil {
		return spec.Eventing.Endpoint
	}
	return ""
}

// returns "default", "custom" or "no" based on args
func dependencyState(url, defaultUrl string) string {
	switch {
	case url == defaultUrl:
		return "default"
	case url == "":
		return "no"
	default:
		return "custom"
	}
}
