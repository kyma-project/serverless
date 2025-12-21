package state

import (
	"context"
	"testing"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnOptionalDependencies(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	tracingCollectorURL := "http://telemetry-otlp-traces.some-ns.svc.cluster.local:4318/v1/traces"
	customEventingURL := "eventing-url"
	configurationReadyMsg := "Configuration ready"

	testCases := map[string]struct {
		tracing               *v1alpha1.Endpoint
		eventing              *v1alpha1.Endpoint
		extraCR               []client.Object
		expectedTracingURL    string
		expectedEventingURL   string
		expectedStatusMessage string
	}{
		"Eventing and tracing is set manually": {
			tracing:               &v1alpha1.Endpoint{Endpoint: tracingCollectorURL},
			eventing:              &v1alpha1.Endpoint{Endpoint: customEventingURL},
			expectedEventingURL:   customEventingURL,
			expectedTracingURL:    tracingCollectorURL,
			expectedStatusMessage: configurationReadyMsg,
		},
		"Tracing is not set, TracePipeline svc is available": {
			extraCR:               []client.Object{fixTracingSvc()},
			eventing:              &v1alpha1.Endpoint{Endpoint: ""},
			expectedTracingURL:    tracingCollectorURL,
			expectedEventingURL:   v1alpha1.EndpointDisabled,
			expectedStatusMessage: configurationReadyMsg,
		},
		"Tracing is not set, TracePipeline svc is not available": {
			expectedEventingURL:   v1alpha1.DefaultEventingEndpoint,
			expectedTracingURL:    v1alpha1.EndpointDisabled,
			expectedStatusMessage: configurationReadyMsg,
		},
		"Tracing and eventing is disabled": {
			tracing:               &v1alpha1.Endpoint{Endpoint: ""},
			eventing:              &v1alpha1.Endpoint{Endpoint: ""},
			expectedEventingURL:   v1alpha1.EndpointDisabled,
			expectedTracingURL:    v1alpha1.EndpointDisabled,
			expectedStatusMessage: configurationReadyMsg,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			s := &systemState{
				instance: v1alpha1.Serverless{
					Spec: v1alpha1.ServerlessSpec{
						Eventing: testCase.eventing,
						Tracing:  testCase.tracing,
					},
				},
				flagsBuilder: flags.NewBuilder(),
			}
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(testCase.extraCR...).Build()
			r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c, EventRecorder: record.NewFakeRecorder(5)}}
			next, result, err := sFnOptionalDependencies(ctx, r, s)
			require.Nil(t, err)
			require.Nil(t, result)
			requireEqualFunc(t, sFnControllerConfiguration, next)

			status := s.instance.Status
			assert.Equal(t, testCase.expectedEventingURL, status.EventingEndpoint)
			assert.Equal(t, testCase.expectedTracingURL, status.TracingEndpoint)
		})
	}

	t.Run("configure chart flags in release if status is up-to date", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Spec: v1alpha1.ServerlessSpec{Eventing: &v1alpha1.Endpoint{Endpoint: customEventingURL}},
				Status: v1alpha1.ServerlessStatus{
					Conditions: []metav1.Condition{
						{
							Type:   string(v1alpha1.ConditionTypeConfigured),
							Status: metav1.ConditionTrue,
						},
					},
					EventingEndpoint: customEventingURL,
					TracingEndpoint:  tracingCollectorURL,
				},
			},
			flagsBuilder: flags.NewBuilder(),
		}
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(fixTracingSvc()).Build()
		r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c}}

		_, _, err := sFnOptionalDependencies(context.Background(), r, s)
		require.NoError(t, err)

		currentFlags, err := s.flagsBuilder.Build()
		require.NoError(t, err)

		overrideURL, found := getFlagByPath(currentFlags, "containers", "manager", "configuration", "data", "functionTraceCollectorEndpoint", "value")
		require.True(t, found)
		assert.Equal(t, tracingCollectorURL, overrideURL)

		overrideURL, found = getFlagByPath(currentFlags, "containers", "manager", "configuration", "data", "functionPublisherProxyAddress", "value")
		require.True(t, found)
		assert.Equal(t, customEventingURL, overrideURL)
	})
}

func fixTracingSvc() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "telemetry-otlp-traces",
			Namespace: "some-ns",
		},
	}
}

func getFlagByPath(flags map[string]interface{}, path ...string) (string, bool) {
	value := flags
	var item interface{}
	for _, pathItem := range path {
		var ok bool
		item, ok = value[pathItem]
		if !ok {
			return "", false
		}
		value, ok = item.(map[string]interface{})
		if !ok {
			break
		}
	}

	out, ok := item.(string)
	return out, ok
}
