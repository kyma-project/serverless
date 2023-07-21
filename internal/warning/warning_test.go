package warning

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuilder_Build(t *testing.T) {
	t.Run("build multiple warnings", func(t *testing.T) {
		warning := NewBuilder().
			With("warn 1").
			With("warn 2").
			Build()

		require.Equal(t, "Warning: warn 1; warn 2", warning)
	})
	t.Run("build empty warning", func(t *testing.T) {
		warning := NewBuilder().Build()
		require.Equal(t, "", warning)
	})
}
