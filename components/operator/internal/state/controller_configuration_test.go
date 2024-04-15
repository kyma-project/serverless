package state

import (
	"context"
	"testing"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	cpuUtilizationTest         = "test-CPU-utilization-percentage"
	requeueDurationTest        = "test-requeue-duration"
	executorArgsTest           = "test-build-executor-args"
	maxSimultaneousJobsTest    = "test-max-simultaneous-jobs"
	healthzLivenessTimeoutTest = "test-healthz-liveness-timeout"
	requestBodyLimitMbTest     = "test-request-body-limit-mb"
	timeoutSecTest             = "test-timeout-sec"
	buildJobPresetTest         = "test=default-build-job-preset"
	runtimePodPresetTest       = "test-default-runtime-pod-preset"
)

func Test_sFnControllerConfiguration(t *testing.T) {
	configurationReadyMsg := "Configuration ready"

	t.Run("update status additional configuration overrides", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				Spec: v1alpha1.DockerRegistrySpec{
					HealthzLivenessTimeout: healthzLivenessTimeoutTest,
				},
			},
			flagsBuilder: chart.NewFlagsBuilder(),
		}

		c := fake.NewClientBuilder().Build()
		eventRecorder := record.NewFakeRecorder(10)
		r := &reconciler{log: zap.NewNop().Sugar(), k8s: k8s{client: c, EventRecorder: eventRecorder}}
		next, result, err := sFnControllerConfiguration(context.TODO(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnApplyResources, next)

		status := s.instance.Status
		require.Equal(t, healthzLivenessTimeoutTest, status.HealthzLivenessTimeout)

		require.Equal(t, v1alpha1.StateProcessing, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			configurationReadyMsg,
		)

		expectedEvents := []string{
			"Normal Configuration Duration of health check set from '' to 'test-healthz-liveness-timeout'",
		}

		for _, expectedEvent := range expectedEvents {
			require.Equal(t, expectedEvent, <-eventRecorder.Events)
		}
	})

	t.Run("reconcile from configurationError", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.DockerRegistry{
				Status: v1alpha1.DockerRegistryStatus{
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
					State: v1alpha1.StateError,
				},
			},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   chart.NewFlagsBuilder(),
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "boo",
			},
		}
		r := &reconciler{
			log: zap.NewNop().Sugar(),
			k8s: k8s{
				client:        fake.NewClientBuilder().WithObjects(secret).Build(),
				EventRecorder: record.NewFakeRecorder(2),
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
