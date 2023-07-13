package state

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
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

func fixEmptyManifestCache() chart.ManifestCache {
	return fixManifestCache("")
}

func fixManifestCache(manifest string) chart.ManifestCache {
	cache := chart.NewInMemoryManifestCache()
	cache.Set(context.Background(), types.NamespacedName{
		Name:      testInstalledServerless.GetName(),
		Namespace: testInstalledServerless.GetNamespace(),
	}, chart.ServerlessSpecManifest{Manifest: manifest})

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

func requireContainsCondition(t *testing.T, status v1alpha1.ServerlessStatus,
	conditionType v1alpha1.ConditionType, conditionStatus metav1.ConditionStatus, conditionReason v1alpha1.ConditionReason, conditionMessage string) {
	hasExpectedCondition := false
	for _, condition := range status.Conditions {
		if condition.Type == string(conditionType) {
			require.Equal(t, string(conditionReason), condition.Reason)
			require.Equal(t, conditionStatus, condition.Status)
			require.Equal(t, conditionMessage, condition.Message)
			hasExpectedCondition = true
		}
	}
	require.True(t, hasExpectedCondition)
}

func fixServerlessClusterWideExternalRegistrySecret() *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "serverless-registry-config",
			Namespace: "kyma-test",
			Labels: map[string]string{
				"serverless.kyma-project.io/remote-registry": "config",
				"serverless.kyma-project.io/config":          "credentials",
			},
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			"registryAddress": []byte("test-registry-address"),
		},
	}
}
