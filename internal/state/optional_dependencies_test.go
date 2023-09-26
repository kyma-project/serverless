package state

import (
	"context"
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	cpuUtilizationTest          = "test-CPU-utilization-percentage"
	requeueDurationTest         = "test-requeue-duration"
	executorArgsTest            = "test-build-executor-args"
	maxSimultaneousJobsTest     = "test-max-simultaneous-jobs"
	healthzLivenessTimeoutTest  = "test-healthz-liveness-timeout"
	requestBodyLimitMbTest      = "test-request-body-limit-mb"
	timeoutSecTest              = "test-timeout-sec"
	defaultBuildJobPresetTest   = "test=default-build-job-preset"
	defaultRuntimePodPresetTest = "test-default-runtime-pod-preset"
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
				chartConfig: &chart.Config{
					Release: chart.Release{
						Flags: chart.EmptyFlags(),
					},
				},
			}
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(testCase.extraCR...).Build()
			r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c, EventRecorder: record.NewFakeRecorder(5)}}
			next, result, err := sFnOptionalDependencies(ctx, r, s)

			expectedNext := sFnApplyResources
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

	t.Run("update status additional configuration overrides", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Spec: v1alpha1.ServerlessSpec{
					TargetCPUUtilizationPercentage:   cpuUtilizationTest,
					FunctionRequeueDuration:          requeueDurationTest,
					FunctionBuildExecutorArgs:        executorArgsTest,
					FunctionBuildMaxSimultaneousJobs: maxSimultaneousJobsTest,
					HealthzLivenessTimeout:           healthzLivenessTimeoutTest,
					FunctionRequestBodyLimitMb:       requestBodyLimitMbTest,
					FunctionTimeoutSec:               timeoutSecTest,
					DefaultBuildJobPreset:            defaultBuildJobPresetTest,
					DefaultRuntimePodPreset:          defaultRuntimePodPresetTest,
				},
			},
			chartConfig: &chart.Config{
				Release: chart.Release{
					Flags: chart.EmptyFlags(),
				},
			},
		}

		c := fake.NewClientBuilder().WithScheme(scheme).Build()
		eventRecorder := record.NewFakeRecorder(10)
		r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c, EventRecorder: eventRecorder}}
		next, result, err := sFnOptionalDependencies(context.TODO(), r, s)

		expectedNext := sFnApplyResources
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)

		status := s.instance.Status
		require.Equal(t, cpuUtilizationTest, status.CPUUtilizationPercentage)
		require.Equal(t, requeueDurationTest, status.RequeueDuration)
		require.Equal(t, executorArgsTest, status.BuildExecutorArgs)
		require.Equal(t, maxSimultaneousJobsTest, status.BuildMaxSimultaneousJobs)
		require.Equal(t, healthzLivenessTimeoutTest, status.HealthzLivenessTimeout)
		require.Equal(t, requestBodyLimitMbTest, status.RequestBodyLimitMb)
		require.Equal(t, timeoutSecTest, status.TimeoutSec)
		require.Equal(t, defaultBuildJobPresetTest, status.DefaultBuildJobPreset)
		require.Equal(t, defaultRuntimePodPresetTest, status.DefaultRuntimePodPreset)

		require.Equal(t, v1alpha1.StateProcessing, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			configurationReadyMsg,
		)

		expectedEvents := []string{
			"Normal Configuration CPU utilization set from '' to 'test-CPU-utilization-percentage'",
			"Normal Configuration Function requeue duration set from '' to 'test-requeue-duration'",
			"Normal Configuration Function build executor args set from '' to 'test-build-executor-args'",
			"Normal Configuration Max number of simultaneous jobs set from '' to 'test-max-simultaneous-jobs'",
			"Normal Configuration Duration of health check set from '' to 'test-healthz-liveness-timeout'",
			"Normal Configuration Max size of request body set from '' to 'test-request-body-limit-mb'",
			"Normal Configuration Timeout set from '' to 'test-timeout-sec'",
			"Normal Configuration Default build job preset set from '' to 'test=default-build-job-preset'",
			"Normal Configuration Default runtime pod preset set from '' to 'test-default-runtime-pod-preset'",
			"Normal Configuration Eventing endpoint set from '' to 'http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish'",
		}

		for _, expectedEvent := range expectedEvents {
			require.Equal(t, expectedEvent, <-eventRecorder.Events)
		}
	})

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
					TracingEndpoint:  v1alpha1.EndpointDisabled,
				},
				Spec: v1alpha1.ServerlessSpec{
					Eventing: &v1alpha1.Endpoint{Endpoint: "test-event-URL"},
					Tracing:  &v1alpha1.Endpoint{Endpoint: v1alpha1.EndpointDisabled},
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal: pointer.Bool(false),
						SecretName:     pointer.String("boo"),
					},
				},
			},
			statusSnapshot: v1alpha1.ServerlessStatus{
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

		expectedNext := sFnApplyResources

		next, result, err := sFnOptionalDependencies(context.Background(), r, s)
		require.Nil(t, result)
		require.NoError(t, err)
		requireEqualFunc(t, expectedNext, next)
		requireContainsCondition(t, s.instance.Status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			configurationReadyMsg)
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

		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(fixTracingSvc()).Build()
		r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c}}

		_, _, err := sFnOptionalDependencies(context.Background(), r, s)
		require.NoError(t, err)

		require.NotNil(t, s.chartConfig)
		overrideURL, found := getFlagByPath(s.chartConfig.Release.Flags, "containers", "manager", "configuration", "data", "functionTraceCollectorEndpoint", "value")
		require.True(t, found)
		assert.Equal(t, tracingCollectorURL, overrideURL)

		overrideURL, found = getFlagByPath(s.chartConfig.Release.Flags, "containers", "manager", "configuration", "data", "functionPublisherProxyAddress", "value")
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
