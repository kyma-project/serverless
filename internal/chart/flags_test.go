package chart

import (
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestAppendInternalRegistryFlags(t *testing.T) {
	t.Run("append internal registry flags", func(t *testing.T) {

		flags := AppendInternalRegistryFlags(map[string]interface{}{}, true)

		require.Equal(t, map[string]interface{}{
			"dockerRegistry": map[string]interface{}{
				"enableInternal": true,
			},
		}, flags)
	})
}

func TestAppendK3dRegistryFlags(t *testing.T) {
	t.Run("append k3d registry flags", func(t *testing.T) {

		flags := AppendK3dRegistryFlags(map[string]interface{}{},
			false,
			v1alpha1.DefaultRegistryAddress,
			v1alpha1.DefaultRegistryAddress,
		)

		require.Equal(t, map[string]interface{}{
			"dockerRegistry": map[string]interface{}{
				"enableInternal":  false,
				"registryAddress": v1alpha1.DefaultRegistryAddress,
				"serverAddress":   v1alpha1.DefaultRegistryAddress,
			},
		}, flags)
	})
}

func TestAppendExternalRegistryFlags(t *testing.T) {
	t.Run("append external registry flags", func(t *testing.T) {

		flags := AppendExternalRegistryFlags(map[string]interface{}{},
			false,
			"username",
			"password",
			"registryAddress",
			"serverAddress",
		)

		require.Equal(t, map[string]interface{}{
			"dockerRegistry": map[string]interface{}{
				"enableInternal":  false,
				"username":        "username",
				"password":        "password",
				"registryAddress": "registryAddress",
				"serverAddress":   "serverAddress",
			},
		}, flags)
	})
}
