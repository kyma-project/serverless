package state

import (
	"context"
	"testing"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/flags"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	requeueDurationTest        = "test-requeue-duration"
	healthzLivenessTimeoutTest = "test-healthz-liveness-timeout"
	runtimePodPresetTest       = "test-default-runtime-pod-preset"
	logLevelTest               = "test-log-level"
	logFormatTest              = "test-log-format"
)

func Test_sFnControllerConfiguration(t *testing.T) {
	configurationReadyMsg := "Configuration ready"

	t.Run("update status in buildless mode", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Spec: v1alpha1.ServerlessSpec{},
			},
			flagsBuilder: flags.NewBuilder(),
		}

		c := fake.NewClientBuilder().WithObjects(
			fixTestNode("node-1"),
			fixTestNode("node-2"),
		).Build()
		eventRecorder := record.NewFakeRecorder(4)
		r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c, EventRecorder: eventRecorder}}
		next, result, err := sFnControllerConfiguration(context.TODO(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnApplyResources, next)

		status := s.instance.Status
		require.Equal(t, slowRuntimePreset, status.DefaultRuntimePodPreset)

		require.Equal(t, v1alpha1.StateProcessing, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			configurationReadyMsg,
		)

		expectedEvents := []string{
			"Normal Configuration Default runtime pod preset set from '' to 'XS'",
		}

		for _, expectedEvent := range expectedEvents {
			require.Equal(t, expectedEvent, <-eventRecorder.Events)
		}
	})

	t.Run("update status with slow defaults", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Spec: v1alpha1.ServerlessSpec{},
			},
			flagsBuilder: flags.NewBuilder(),
		}

		c := fake.NewClientBuilder().WithObjects(
			fixTestNode("node-1"),
			fixTestNode("node-2"),
		).Build()
		eventRecorder := record.NewFakeRecorder(4)
		r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c, EventRecorder: eventRecorder}}
		next, result, err := sFnControllerConfiguration(context.TODO(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnApplyResources, next)

		status := s.instance.Status
		require.Equal(t, slowRuntimePreset, status.DefaultRuntimePodPreset)

		require.Equal(t, v1alpha1.StateProcessing, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			configurationReadyMsg,
		)

		expectedEvents := []string{
			"Normal Configuration Default runtime pod preset set from '' to 'XS'",
			"Normal Configuration Log level set from '' to 'info'",
			"Normal Configuration Log format set from '' to 'json'",
		}

		for _, expectedEvent := range expectedEvents {
			require.Equal(t, expectedEvent, <-eventRecorder.Events)
		}
	})

	t.Run("update slow default to normal ones", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Spec: v1alpha1.ServerlessSpec{},
				Status: v1alpha1.ServerlessStatus{
					DefaultRuntimePodPreset: slowRuntimePreset,
				},
			},
			flagsBuilder: flags.NewBuilder(),
		}

		c := fake.NewClientBuilder().WithObjects(
			fixTestNode("node-1"),
			fixTestNode("node-2"),
			fixTestNode("node-3"),
			fixTestNode("node-4"),
		).Build()
		eventRecorder := record.NewFakeRecorder(4)
		r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c, EventRecorder: eventRecorder}}
		next, result, err := sFnControllerConfiguration(context.TODO(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnApplyResources, next)

		status := s.instance.Status
		require.Equal(t, largeRuntimePreset, status.DefaultRuntimePodPreset)

		require.Equal(t, v1alpha1.StateProcessing, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			configurationReadyMsg,
		)

		expectedEvents := []string{
			"Normal Configuration Default runtime pod preset set from 'XS' to 'L'",
			"Normal Configuration Log level set from '' to 'info'",
			"Normal Configuration Log format set from '' to 'json'",
		}

		for _, expectedEvent := range expectedEvents {
			require.Equal(t, expectedEvent, <-eventRecorder.Events)
		}
	})

	t.Run("update status additional configuration overrides", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Spec: v1alpha1.ServerlessSpec{
					FunctionRequeueDuration: requeueDurationTest,
					HealthzLivenessTimeout:  healthzLivenessTimeoutTest,
					DefaultRuntimePodPreset: runtimePodPresetTest,
					LogLevel:                logLevelTest,
					LogFormat:               logFormatTest,
				},
			},
			flagsBuilder: flags.NewBuilder(),
		}

		c := fake.NewClientBuilder().Build()
		eventRecorder := record.NewFakeRecorder(10)
		r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c, EventRecorder: eventRecorder}}
		next, result, err := sFnControllerConfiguration(context.TODO(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnApplyResources, next)

		status := s.instance.Status
		require.Equal(t, requeueDurationTest, status.RequeueDuration)
		require.Equal(t, healthzLivenessTimeoutTest, status.HealthzLivenessTimeout)
		require.Equal(t, runtimePodPresetTest, status.DefaultRuntimePodPreset)
		require.Equal(t, logLevelTest, status.LogLevel)
		require.Equal(t, logFormatTest, status.LogFormat)

		require.Equal(t, v1alpha1.StateProcessing, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			configurationReadyMsg,
		)

		expectedEvents := []string{
			"Normal Configuration Function requeue duration set from '' to 'test-requeue-duration'",
			"Normal Configuration Duration of health check set from '' to 'test-healthz-liveness-timeout'",
			"Normal Configuration Default runtime pod preset set from '' to 'test-default-runtime-pod-preset'",
			"Normal Configuration Log level set from '' to 'test-log-level'",
			"Normal Configuration Log format set from '' to 'test-log-format'",
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
				},
			},
			statusSnapshot: v1alpha1.ServerlessStatus{},
			flagsBuilder:   flags.NewBuilder(),
		}
		r := &reconciler{
			log: zap.NewNop().Sugar(),
			k8s: k8s{
				client:        fake.NewClientBuilder().Build(),
				EventRecorder: record.NewFakeRecorder(4),
			},
		}

		next, result, err := sFnControllerConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnApplyResources, next)
		requireContainsCondition(t, s.instance.Status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			configurationReadyMsg)
		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
	})
}

func fixTestNode(name string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}
