package gitrepository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCleanup(t *testing.T) {
	t.Run("remove crd", func(t *testing.T) {
		ctx := context.Background()
		c := fake.NewClientBuilder().
			WithScheme(apiextensionsscheme.Scheme).
			WithObjects(fixGitRepoCRD()).
			Build()

		err := Cleanup(ctx, c)

		require.NoError(t, err)

		err = c.Get(ctx, types.NamespacedName{
			Name: gitRepoCRDName,
		}, fixGitRepoCRD())
		require.True(t, errors.IsNotFound(err))
	})

	t.Run("crd not found", func(t *testing.T) {
		ctx := context.Background()
		c := fake.NewClientBuilder().
			WithScheme(apiextensionsscheme.Scheme).
			Build()

		err := Cleanup(ctx, c)

		require.NoError(t, err)
	})

	t.Run("client get error", func(t *testing.T) {
		ctx := context.Background()
		c := fake.NewClientBuilder().Build()

		err := Cleanup(ctx, c)

		require.Error(t, err)
	})
}

func fixGitRepoCRD() *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: v1.ObjectMeta{
			Name: gitRepoCRDName,
		},
	}
}
