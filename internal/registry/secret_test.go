package registry

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestListExternalRegistrySecrets(t *testing.T) {

	t.Run("returns error when external registry secret not found", func(t *testing.T) {
		err := ListExternalRegistrySecrets()
		require.Error(t, err)
	})
}
