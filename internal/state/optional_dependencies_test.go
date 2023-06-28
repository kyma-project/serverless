package state

import (
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

		stateFn := sFnOptionalDependencies()
		next, result, err := stateFn(nil, nil, s)

		expectedNext := sFnUpdateStatusAndRequeue
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)

		status := s.instance.Status
		require.Equal(t, "test-event-URL", status.EventingEndpoint)
		require.Equal(t, "test-trace-URL", status.TracingEndpoint)

		require.Equal(t, v1alpha1.StateProcessing, status.State)
		require.Len(t, status.Conditions, 1)
		condition := status.Conditions[0]
		require.Equal(t, string(v1alpha1.ConditionTypeConfigured), condition.Type)
		require.Equal(t, string(v1alpha1.ConditionReasonConfigured), condition.Reason)
		require.Equal(t, "Configured with custom Publisher Proxy URL and custom Trace Collector URL.", condition.Message)
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

		stateFn := sFnOptionalDependencies()
		next, result, err := stateFn(nil, nil, s)

		expectedNext := sFnApplyResources()

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})
}
