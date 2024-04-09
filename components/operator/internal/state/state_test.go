package state

import (
	"context"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/chart"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

var (
	testInstalledDockerRegistry = v1alpha1.DockerRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.DockerRegistrySpec{
			DockerRegistry: &v1alpha1.DockerRegistryCfg{
				EnableInternal: ptr.To[bool](false),
			},
		},
		Status: v1alpha1.DockerRegistryStatus{
			Conditions: []metav1.Condition{
				{
					Type:   string(v1alpha1.ConditionTypeConfigured),
					Status: metav1.ConditionTrue,
					Reason: string(v1alpha1.ConditionReasonConfiguration),
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
	return fixManifestCache("---")
}

func fixManifestCache(manifest string) chart.ManifestCache {
	cache := chart.NewInMemoryManifestCache()
	_ = cache.Set(context.Background(), types.NamespacedName{
		Name:      testInstalledDockerRegistry.GetName(),
		Namespace: testInstalledDockerRegistry.GetNamespace(),
	}, chart.ServerlessSpecManifest{Manifest: manifest, CustomFlags: map[string]interface{}{}})

	return cache
}

func requireEqualFunc(t *testing.T, expected, actual stateFn) {
	require.NotNil(t, actual)

	expectedFnName := getFnName(expected)
	actualFnName := getFnName(actual)

	if expectedFnName == actualFnName {
		// return if functions are simply same
		return
	}

	expectedElems := strings.Split(expectedFnName, "/")
	actualElems := strings.Split(actualFnName, "/")

	// check package paths (prefix)
	require.Equal(t,
		strings.Join(expectedElems[0:len(expectedElems)-2], "/"),
		strings.Join(actualElems[0:len(actualElems)-2], "/"),
	)

	// check direct fn names (suffix)
	require.Equal(t,
		getDirectFnName(expectedElems[len(expectedElems)-1]),
		getDirectFnName(actualElems[len(actualElems)-1]),
	)
}

func getDirectFnName(nameSuffix string) string {
	elements := strings.Split(nameSuffix, ".")
	for i := range elements {
		elemI := len(elements) - i - 1
		if !strings.HasPrefix(elements[elemI], "func") {
			return elements[elemI]
		}
	}

	return ""
}

func getFnName(fn stateFn) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

func requireContainsCondition(t *testing.T, status v1alpha1.DockerRegistryStatus,
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
