package state

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_sFnOptionalDependencies(t *testing.T) {
	t.Run("update status with endpoints info", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Spec: v1alpha1.ServerlessSpec{
					Eventing: &v1alpha1.Endpoint{Endpoint: "test-event-URL"},
					Tracing:  &v1alpha1.Endpoint{Endpoint: "test-trace-URL"},
				},
			},
		}

		next, result, err := sFnOptionalDependencies(nil, nil, s)

		expectedNext := sFnUpdateStatusAndRequeue
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)

		status := s.instance.Status
		require.Equal(t, "test-event-URL", status.EventingEndpoint)
		require.Equal(t, "test-trace-URL", status.TracingEndpoint)

		require.Equal(t, v1alpha1.StateProcessing, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConfigured,
			"Configured with custom Publisher Proxy URL and custom Trace Collector URL.",
		)
	})

	t.Run("next state", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Spec: v1alpha1.ServerlessSpec{
					Eventing: &v1alpha1.Endpoint{Endpoint: "test-event-URL"},
					Tracing:  &v1alpha1.Endpoint{Endpoint: v1alpha1.DefaultTraceCollectorURL},
				},
				Status: v1alpha1.ServerlessStatus{
					Conditions: []metav1.Condition{
						{
							Type:   string(v1alpha1.ConditionTypeConfigured),
							Status: metav1.ConditionTrue,
						},
					},
					EventingEndpoint: "test-event-URL",
					TracingEndpoint:  v1alpha1.DefaultTraceCollectorURL,
				},
			},
			snapshot: v1alpha1.ServerlessStatus{
				EventingEndpoint: "test-event-URL",
				TracingEndpoint:  v1alpha1.DefaultTraceCollectorURL,
			},
			chartConfig: &chart.Config{
				Release: chart.Release{
					Flags: map[string]interface{}{},
				},
			},
		}

		next, result, err := sFnOptionalDependencies(nil, nil, s)

		expectedNext := sFnApplyResources
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
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
					TracingEndpoint:  v1alpha1.DefaultTraceCollectorURL,
				},
				Spec: v1alpha1.ServerlessSpec{
					Eventing: &v1alpha1.Endpoint{Endpoint: "test-event-URL"},
					Tracing:  &v1alpha1.Endpoint{Endpoint: v1alpha1.DefaultTraceCollectorURL},
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
			"Configured with custom Publisher Proxy URL and default Trace Collector URL.")
		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
	})
}
