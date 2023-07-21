package warning

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuilder_Build(t *testing.T) {
	t.Run("build multiple warnings", func(t *testing.T) {
		builder := NewBuilder()
		builder.With("warn 1")
		builder.With("warn 2")

		require.Equal(t, "Warning: warn 1; warn 2", builder.Build())
	})
}
