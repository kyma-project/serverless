package git

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetLatestCommit_CommitSHA(t *testing.T) {
	t.Run("returns commit SHA directly without remote call", func(t *testing.T) {
		sha := "25ea1d5577a4362500fc77d4ea9a2bfeb3665c05"
		result, err := GetLatestCommit("http://should-not-be-called", sha, nil)
		require.NoError(t, err)
		require.Equal(t, sha, result)
	})

	t.Run("does not treat short hex string as commit SHA", func(t *testing.T) {
		// A 7-char abbreviated SHA should NOT be treated as a full commit SHA —
		// it would fail on an unreachable URL just like a branch name would.
		_, err := GetLatestCommit("http://unreachable.invalid", "25ea1d5", nil)
		require.Error(t, err)
	})

	t.Run("does not treat non-hex string as commit SHA", func(t *testing.T) {
		_, err := GetLatestCommit("http://unreachable.invalid", "main", nil)
		require.Error(t, err)
	})
}
