package annotation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testDeployCR = &appsv1.Deployment{
		TypeMeta: v1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-deploy",
			Namespace: "default",
			Annotations: map[string]string{
				"reconciler.kyma-project.io/managed-by-reconciler-disclaimer": "test message",
			},
		},
	}
)

func TestAddDoNotEditDisclaimer(t *testing.T) {
	t.Run("add disclaimer", func(t *testing.T) {
		obj := unstructured.Unstructured{}
		obj = AddDoNotEditDisclaimer(obj)

		require.Equal(t, message, obj.GetAnnotations()[annotation])
	})
}

func TestDeleteReconcilerDisclaimer(t *testing.T) {
	t.Run("prepare good request", func(t *testing.T) {
		mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{})
		mapper.Add(appsv1.SchemeGroupVersion.WithKind("Deployment"), meta.RESTScopeNamespace)

		server := fixTestPatchServer(t)
		defer server.Close()

		client := fake.NewClientBuilder().
			WithRESTMapper(mapper).
			Build()

		obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(testDeployCR)
		require.NoError(t, err)

		err = DeleteReconcilerDisclaimer(client, rest.Config{
			Host: server.URL,
		}, unstructured.Unstructured{Object: obj})
		require.NoError(t, err)
	})
}

func fixTestPatchServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "PATCH", r.Method)
		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, reconcilerPatch, string(bodyBytes))

		jsonDeploy, err := json.Marshal(testDeployCR)
		require.NoError(t, err)

		w.Header().Add("Content-Type", "application/json")
		fmt.Fprint(w, string(jsonDeploy))
	}))
}
