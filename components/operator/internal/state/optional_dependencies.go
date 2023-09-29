package state

import (
	"context"

	//TODO: change to "github.com/kyma-project/serverless-manager/components/operator/api/v1alpha1" after new SO structure will be on github
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	//TODO: change to "github.com/kyma-project/serverless-manager/components/operator/internal/registry" after new SO structure will be on github
	"github.com/kyma-project/serverless-manager/internal/chart"
	//TODO: change to "github.com/kyma-project/serverless-manager/components/operator/internal/tracing" after new SO structure will be on github
	"github.com/kyma-project/serverless-manager/internal/tracing"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// enable or disable serverless optional dependencies based on the Serverless Spec and installed module on the cluster
func sFnOptionalDependencies(ctx context.Context, r *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
	// TODO: add functionality of auto-detecting these dependencies by checking Eventing CRs if user does not override these values.
	// checking these URLs manually is not possible because of lack of istio-sidecar in the serverless-operator

	tracingURL, err := getTracingURL(ctx, r.client, s.instance.Spec)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while fetching tracing URL")
	}
	eventingURL := getEventingURL(s.instance.Spec)

	updateStatus(r.k8s, &s.instance, eventingURL, tracingURL)
	s.setState(v1alpha1.StateProcessing)
	s.instance.UpdateConditionTrue(
		v1alpha1.ConditionTypeConfigured,
		v1alpha1.ConditionReasonConfigured,
		"Configuration ready",
	)

	configureDependenciesFlags(s)

	return nextState(sFnApplyResources)
}

func getTracingURL(ctx context.Context, client client.Client, spec v1alpha1.ServerlessSpec) (string, error) {
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

func updateStatus(eventRecorder record.EventRecorder, instance *v1alpha1.Serverless, eventingURL, tracingURL string) {
	spec := instance.Spec

	fields := []struct {
		specField   string
		statusField *string
		fieldName   string
	}{
		{spec.TargetCPUUtilizationPercentage, &instance.Status.CPUUtilizationPercentage, "CPU utilization"},
		{spec.FunctionRequeueDuration, &instance.Status.RequeueDuration, "Function requeue duration"},
		{spec.FunctionBuildExecutorArgs, &instance.Status.BuildExecutorArgs, "Function build executor args"},
		{spec.FunctionBuildMaxSimultaneousJobs, &instance.Status.BuildMaxSimultaneousJobs, "Max number of simultaneous jobs"},
		{spec.HealthzLivenessTimeout, &instance.Status.HealthzLivenessTimeout, "Duration of health check"},
		{spec.FunctionRequestBodyLimitMb, &instance.Status.RequestBodyLimitMb, "Max size of request body"},
		{spec.FunctionTimeoutSec, &instance.Status.TimeoutSec, "Timeout"},
		{spec.DefaultBuildJobPreset, &instance.Status.DefaultBuildJobPreset, "Default build job preset"},
		{spec.DefaultRuntimePodPreset, &instance.Status.DefaultRuntimePodPreset, "Default runtime pod preset"},
		{eventingURL, &instance.Status.EventingEndpoint, "Eventing endpoint"},
		{tracingURL, &instance.Status.TracingEndpoint, "Tracing endpoint"},
	}

	for _, field := range fields {
		if field.specField != *field.statusField {
			oldStatusValue := *field.statusField
			*field.statusField = field.specField
			eventRecorder.Eventf(
				instance,
				"Normal",
				string(v1alpha1.ConditionReasonConfiguration),
				"%s set from '%s' to '%s'",
				field.fieldName,
				oldStatusValue,
				field.specField,
			)
		}
	}
}

func configureDependenciesFlags(s *systemState) {
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
		WithOptionalDependencies(
			s.instance.Status.EventingEndpoint,
			s.instance.Status.TracingEndpoint,
		).
		WithDefaultPresetFlags(
			s.instance.Status.DefaultBuildJobPreset,
			s.instance.Status.DefaultRuntimePodPreset,
		)
}
