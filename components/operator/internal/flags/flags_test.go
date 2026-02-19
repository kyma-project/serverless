package flags

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithFipsModeEnabled(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		fb := NewBuilder()
		fb.WithFipsModeEnabled(true)

		flagsMap, err := fb.Build()
		require.NoError(t, err)

		expected := map[string]interface{}{
			"containers": map[string]interface{}{
				"manager": map[string]interface{}{
					"fipsModeEnabled": true,
				},
			},
		}

		require.Equal(t, expected, flagsMap)
	})

	t.Run("false", func(t *testing.T) {
		fb := NewBuilder()
		fb.WithFipsModeEnabled(false)

		flagsMap, err := fb.Build()
		require.NoError(t, err)

		expected := map[string]interface{}{
			"containers": map[string]interface{}{
				"manager": map[string]interface{}{
					"fipsModeEnabled": false,
				},
			},
		}

		require.Equal(t, expected, flagsMap)
	})
}
