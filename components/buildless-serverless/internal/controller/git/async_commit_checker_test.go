package git

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAsyncLastCommitChecker(t *testing.T) {
	t.Run("order last commit check and cleanup cache entry", func(t *testing.T) {
		id := "order-id"
		repo := "test-repo"
		ref := "test-ref"
		auth := &GitAuth{
			secretName:      "test",
			secretNamespace: "default",
		}
		checker := asyncLatestCommitChecker{
			ctx: context.Background(),
			log: zap.NewNop().Sugar(),
			getLatestCommit: func(repo, ref string, auth *GitAuth) (string, error) {
				return "test-commit", nil
			},
		}

		result := checker.CollectOrder(id)
		require.Nil(t, result)

		checker.MakeOrder(id, repo, ref, auth)

		time.Sleep(2 * time.Millisecond) // wait for async operation to complete

		result = checker.CollectOrder(id)
		require.NotNil(t, result, "commit check should be ordered and finished")
		require.Equal(t, "test-commit", result.Commit)
		require.NoError(t, result.Error)

		// subsequent call should return nil as the entry should be removed from cache
		result = checker.CollectOrder(id)
		require.Nil(t, result, "commit check order should be removed from cache after collecting the result")

	})
}
