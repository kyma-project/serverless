package state

import (
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/controller/fsm"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// requireEqualFunc compares two stateFns based on their names returned from the reflect package
// names returned from the package may be different for any go/dlv compiler version
// for go1.22 returned name is in format:
// github.com/kyma-project/keda-manager/pkg/reconciler.Test_sFnServedFilter.func4.sFnUpdateStatus.3
// we assume that function name starts with "sFn"
func requireEqualFunc(t *testing.T, expected, actual fsm.StateFn) {
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
	// try to find something starts with "sFn" on the beginning or immediately after the dot
	nameMatch := regexp.MustCompile(`(^|\.)(sFn[^.]+)`).FindStringSubmatch(nameSuffix)
	if len(nameMatch) != 3 {
		return nameSuffix
	}
	return nameMatch[2]
}

func getFnName(fn fsm.StateFn) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

func requireContainsCondition(t *testing.T, status serverlessv1alpha2.FunctionStatus,
	conditionType serverlessv1alpha2.ConditionType, conditionStatus metav1.ConditionStatus,
	conditionReason serverlessv1alpha2.ConditionReason, conditionMessage string) {
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
