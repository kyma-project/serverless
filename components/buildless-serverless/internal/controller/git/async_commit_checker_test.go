package git

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_AsyncLatestCommitChecker(t *testing.T) {
	t.Run("order last commit check and cleanup cache entry", func(t *testing.T) {
		id := "order-id"
		repo := "test-repo"
		ref := "test-ref"
		auth := &GitAuth{
			secretName:      "test",
			secretNamespace: "default",
		}
		checker := asyncLatestCommitChecker{
			ctx:               context.Background(),
			log:               zap.NewNop().Sugar(),
			cacheElemLifetime: 0,
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

		_, exists := checker.cache.Load(id)
		require.False(t, exists, "cache entry should be removed after collecting the result")
	})

	t.Run("get last order without removing it from cache", func(t *testing.T) {
		id := "order-id"
		order := &OrderResult{
			Commit:    "test-commit",
			timestamp: time.Now(),
		}

		checker := asyncLatestCommitChecker{
			ctx:               context.Background(),
			log:               zap.NewNop().Sugar(),
			cacheElemLifetime: time.Hour,
			getLatestCommit: func(repo, ref string, auth *GitAuth) (string, error) {
				return "test-commit", nil
			},
		}

		checker.cache.Store(id, order)

		result := checker.CollectOrder(id)
		require.NotNil(t, result, "should get existing order result")
		require.Equal(t, "test-commit", result.Commit)

		_, exists := checker.cache.Load(id)
		require.True(t, exists, "cache entry should still exist after collecting the result")
	})

	t.Run("do not order last commit check again if already ordered", func(t *testing.T) {
		id := "order-id"
		repo := "test-repo"
		ref := "test-ref"
		auth := &GitAuth{
			secretName:      "test",
			secretNamespace: "default",
		}

		ordersCount := 0
		checker := asyncLatestCommitChecker{
			ctx: context.Background(),
			log: zap.NewNop().Sugar(),
			getLatestCommit: func(repo, ref string, auth *GitAuth) (string, error) {
				ordersCount++
				time.Sleep(time.Second)
				return "test-commit", nil
			},
		}

		checker.MakeOrder(id, repo, ref, auth)
		checker.MakeOrder(id, repo, ref, auth)
		checker.MakeOrder(id, repo, ref, auth)

		// wait for async operation to complete
		time.Sleep(time.Millisecond * 10)

		require.Equal(t, 1, ordersCount, "commit check should be ordered only once")
	})
}

func Test_clearCacheEvery(t *testing.T) {
	t.Run("remove old entries from cache", func(t *testing.T) {
		id := "order-id"

		checker := asyncLatestCommitChecker{
			ctx: context.Background(),
			log: zap.NewNop().Sugar(),
		}

		// add the entry back to cache to simulate old entry
		checker.cache.Store(id, nil)

		// start cache cleanup with short interval
		checker.clearCacheEvery(time.Millisecond)

		time.Sleep(5 * time.Millisecond) // wait for cache cleanup to run

		_, exists := checker.cache.Load(id)
		require.False(t, exists, "old cache entry should be removed")
	})

	t.Run("stop cache cleanup when context is done", func(t *testing.T) {
		id := "order-id"
		ctx, cancel := context.WithCancel(context.Background())

		checker := asyncLatestCommitChecker{
			ctx: ctx,
			log: zap.NewNop().Sugar(),
		}

		// add the entry back to cache to simulate old entry
		checker.cache.Store(id, nil)

		// start cache cleanup with short interval
		checker.clearCacheEvery(time.Minute)

		cancel()                         // cancel the context to stop the cleanup goroutine
		time.Sleep(5 * time.Millisecond) // wait to ensure goroutine has time to exit

		_, exists := checker.cache.Load(id)
		require.True(t, exists, "entry should still exist as cleanup should be stopped")
	})
}
