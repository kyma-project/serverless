package state

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
)

var (
	testInstalledServerless = v1alpha1.Serverless{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.ServerlessSpec{
			DockerRegistry: &v1alpha1.DockerRegistry{
				EnableInternal: pointer.Bool(false),
			},
		},
		Status: v1alpha1.ServerlessStatus{
			Conditions: []metav1.Condition{
				{
					Type:   string(v1alpha1.ConditionTypeConfigured),
					Status: metav1.ConditionTrue,
					Reason: string(v1alpha1.ConditionReasonConfigurationCheck),
				},
				{
					Type:   string(v1alpha1.ConditionTypeInstalled),
					Status: metav1.ConditionTrue,
					Reason: string(v1alpha1.ConditionReasonInstallation),
				},
			},
			State: v1alpha1.StateReady,
		},
	}
)

func testEmptyManifestCache() chart.ManifestCache {
	cache := chart.NewInMemoryManifestCache()
	cache.Set(context.Background(), types.NamespacedName{
		Name:      testInstalledServerless.GetName(),
		Namespace: testInstalledServerless.GetNamespace(),
	}, nil, "")

	return cache
}

func requireEqualFunc(t *testing.T, expected, actual stateFn) {
	expectedValueOf := reflect.ValueOf(expected)
	actualValueOf := reflect.ValueOf(actual)
	require.True(t, expectedValueOf.Pointer() == actualValueOf.Pointer(),
		fmt.Sprintf("expected '%s', got '%s", getFnName(expected), getFnName(actual)))
}

func getFnName(fn stateFn) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}
