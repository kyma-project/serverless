package state

import (
	"context"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/kyma-project/manager-toolkit/installation/chart"
	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

var (
	testInstalledServerless = v1alpha1.Serverless{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.ServerlessSpec{
			DockerRegistry: &v1alpha1.DockerRegistry{
				EnableInternal: ptr.To[bool](false),
			},
		},
		Status: v1alpha1.ServerlessStatus{
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
		Name:      testInstalledServerless.GetName(),
		Namespace: testInstalledServerless.GetNamespace(),
	}, chart.ContextManifest{Manifest: manifest, CustomFlags: map[string]interface{}{
		"global": map[string]interface{}{
			"commonLabels": map[string]interface{}{
				"app.kubernetes.io/managed-by": "serverless-operator",
			},
		},
		"containers": map[string]interface{}{
			"manager": map[string]interface{}{
				"fipsModeEnabled": false,
			},
		},
	}})

	return cache
}

// requireEqualFunc compares two stateFns based on their names returned from the reflect package
// names returned from the package may be different for any go/dlv compiler version
// for go1.22 returned name is in format:
// github.com/kyma-project/keda-manager/pkg/reconciler.Test_sFnServedFilter.func4.sFnUpdateStatus.3
func requireEqualFunc(t *testing.T, expected, actual stateFn) {
	expectedFnName := getFnName(expected)
	actualFnName := getFnName(actual)

	if expectedFnName == actualFnName {
		// return if functions are simply same
		return
	}

	expectedElems := strings.Split(expectedFnName, "/")
	actualElems := strings.Split(actualFnName, "/")

	// check package paths (prefix)
	// e.g. 'github.com/kyma-project/keda-manager/pkg'
	require.Equal(t,
		strings.Join(expectedElems[0:len(expectedElems)-2], "/"),
		strings.Join(actualElems[0:len(actualElems)-2], "/"),
	)

	// check direct fn names (suffix)
	// e.g. 'reconciler.Test_sFnServedFilter.func4.sFnUpdateStatus.3'
	require.Equal(t,
		getDirectFnName(expectedElems[len(expectedElems)-1]),
		getDirectFnName(actualElems[len(actualElems)-1]),
	)
}

func getDirectFnName(nameSuffix string) string {
	elements := strings.Split(nameSuffix, ".")
	return elements[len(elements)-2]
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
