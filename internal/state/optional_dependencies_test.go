package state

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_sFnOptionalDependencies(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	tracingCollectorURL := "http://telemetry-otlp-traces.some-ns.svc.cluster.local:4318/v1/traces"
	customEventingURL := "eventing-url"

	configuredMsg := "Configured with custom Publisher Proxy URL and custom Trace Collector URL."
	noConfigurationMsg := "Configured with no Publisher Proxy URL and no Trace Collector URL."
	traceConfiguredMsg := "Configured with no Publisher Proxy URL and custom Trace Collector URL."

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
			expectedStatusMessage: configuredMsg,
		},
		"Tracing is not set, TracePipeline svc is available": {
			extraCR:               []client.Object{fixTracingSvc()},
			eventing:              &v1alpha1.Endpoint{Endpoint: ""},
			expectedTracingURL:    tracingCollectorURL,
			expectedEventingURL:   v1alpha1.FeatureDisabled,
			expectedStatusMessage: traceConfiguredMsg,
		},
		"Tracing is not set, TracePipeline svc is not available": {
			expectedEventingURL:   v1alpha1.FeatureDisabled,
			expectedTracingURL:    v1alpha1.FeatureDisabled,
			expectedStatusMessage: noConfigurationMsg,
		},
		"Tracing and eventing is disabled": {
			tracing:               &v1alpha1.Endpoint{Endpoint: ""},
			eventing:              &v1alpha1.Endpoint{Endpoint: ""},
			expectedEventingURL:   v1alpha1.FeatureDisabled,
			expectedTracingURL:    v1alpha1.FeatureDisabled,
			expectedStatusMessage: noConfigurationMsg,
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
			}
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(testCase.extraCR...).Build()
			r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c}}
			next, result, err := sFnOptionalDependencies(ctx, r, s)

			expectedNext := sFnUpdateStatusAndRequeue
			requireEqualFunc(t, expectedNext, next)
			require.Nil(t, result)
			require.Nil(t, err)

			status := s.instance.Status
			assert.Equal(t, testCase.expectedEventingURL, status.EventingEndpoint)
			assert.Equal(t, testCase.expectedTracingURL, status.TracingEndpoint)

			assert.Equal(t, v1alpha1.StateProcessing, status.State)
			requireContainsCondition(t, status,
				v1alpha1.ConditionTypeConfigured,
				metav1.ConditionTrue,
				v1alpha1.ConditionReasonConfigured,
				testCase.expectedStatusMessage,
			)
		})
	}

	t.Run("reconcile from configurationError", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Status: v1alpha1.ServerlessStatus{
					Conditions: []metav1.Condition{
						{
							Type:   string(v1alpha1.ConditionTypeConfigured),
							Status: metav1.ConditionFalse,
							Reason: string(v1alpha1.ConditionReasonConfigurationErr),
						},
						{
							Type:   string(v1alpha1.ConditionTypeInstalled),
							Status: metav1.ConditionTrue,
							Reason: string(v1alpha1.ConditionReasonInstallation),
						},
					},
					State:            v1alpha1.StateError,
					EventingEndpoint: "test-event-URL",
					TracingEndpoint:  v1alpha1.FeatureDisabled,
				},
				Spec: v1alpha1.ServerlessSpec{
					Eventing: &v1alpha1.Endpoint{Endpoint: "test-event-URL"},
					Tracing:  &v1alpha1.Endpoint{Endpoint: v1alpha1.FeatureDisabled},
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal: pointer.Bool(false),
						SecretName:     pointer.String("boo"),
					},
				},
			},
			snapshot: v1alpha1.ServerlessStatus{
				DockerRegistry: "",
			},
			chartConfig: &chart.Config{
				Release: chart.Release{
					Flags: chart.EmptyFlags(),
				},
			},
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "boo",
			},
		}
		r := &reconciler{
			log: zap.NewNop().Sugar(),
			k8s: k8s{
				client: fake.NewClientBuilder().WithObjects(secret).Build(),
			},
		}

		expectedNext := sFnUpdateStatusAndRequeue

		next, result, err := sFnOptionalDependencies(context.Background(), r, s)
		require.Nil(t, result)
		require.NoError(t, err)
		requireEqualFunc(t, expectedNext, next)
		requireContainsCondition(t, s.instance.Status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			"Configured with custom Publisher Proxy URL and no Trace Collector URL.")
		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
	})

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
			chartConfig: &chart.Config{
				Release: chart.Release{
					Flags: map[string]interface{}{},
				},
			},
		}

		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(fixTracingSvc()).Build()
		r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: client}}

		_, _, err := sFnOptionalDependencies(context.Background(), r, s)
		require.NoError(t, err)

		require.NotNil(t, s.chartConfig)
		overrideURL, found := getFlagByPath(s.chartConfig.Release.Flags, "containers", "manager", "envs", "functionTraceCollectorEndpoint", "value")
		require.True(t, found)
		assert.Equal(t, tracingCollectorURL, overrideURL)

		overrideURL, found = getFlagByPath(s.chartConfig.Release.Flags, "containers", "manager", "envs", "functionPublisherProxyAddress", "value")
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
