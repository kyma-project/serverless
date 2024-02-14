package state

import (
	"context"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/tracing"
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
		wrappedErr := errors.Wrap(err, "while fetching tracing URL")
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigurationErr,
			wrappedErr,
		)
		return nil, nil, wrappedErr
	}
	eventingURL := getEventingURL(s.instance.Spec)

	updateOptionalDependenciesStatus(r.k8s, &s.instance, eventingURL, tracingURL)
	configureOptionalDependenciesFlags(s)

	return nextState(sFnControllerConfiguration)
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

func updateOptionalDependenciesStatus(eventRecorder record.EventRecorder, instance *v1alpha1.Serverless, eventingURL, tracingURL string) {
	fields := fieldsToUpdate{
		{eventingURL, &instance.Status.EventingEndpoint, "Eventing endpoint", ""},
		{tracingURL, &instance.Status.TracingEndpoint, "Tracing endpoint", ""},
	}

	updateStatusFields(eventRecorder, instance, fields)
}

func configureOptionalDependenciesFlags(s *systemState) {
	s.flagsBuilder.
		WithOptionalDependencies(
			s.instance.Status.EventingEndpoint,
			s.instance.Status.TracingEndpoint,
		)
}
