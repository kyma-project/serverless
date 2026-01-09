package git

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

//go:generate mockery --name=AsyncLastCommitChecker --output=automock --outpkg=automock --case=underscore
type AsyncLastCommitChecker interface {
	OrderLastCommitCheck(context.Context, string, string, *GitAuth)
	IsLastCommitCheckOrdered(string, string, *GitAuth) bool
	DeleteLastCommitCheckOrder(string, string, *GitAuth)
	GetLastCommitCheckResult(string, string, *GitAuth) *OrderResult
}

type asyncLastCommitChecker struct {
	cache              sync.Map
	log                *zap.SugaredLogger
	cacheEntryLifespan time.Duration
}

type orderCacheKey struct {
	repo string
	ref  string
	auth *GitAuth
}

type OrderResult struct {
	Commit string
	Error  error

	// cancel function to stop order go routine
	cancel context.CancelFunc
}

func NewAsyncLastCommitChecker(log *zap.SugaredLogger, cacheEntryLifespan time.Duration) AsyncLastCommitChecker {
	return &asyncLastCommitChecker{
		log:                log,
		cacheEntryLifespan: cacheEntryLifespan,
	}
}

// OrderLastCommitCheck orders asynchronous git last commit check
func (c *asyncLastCommitChecker) OrderLastCommitCheck(ctx context.Context, repo, ref string, auth *GitAuth) {
	key := orderCacheKey{
		repo: repo,
		ref:  ref,
		auth: auth,
	}

	// iniy empty result to mark that the check has been ordered
	c.cache.Store(key, nil)

	go func() {
		c.log.Debugf("starting async last commit check for %s %s", key.repo, key.ref)
		commit, err := GetLatestCommit(key.repo, key.ref, key.auth)

		// timeout context will be used to cleanup cache entry after some time
		timeoutCtx, cancel := context.WithTimeout(ctx, c.cacheEntryLifespan)
		defer cancel()

		c.log.Debugf("finished async last commit check for %s %s with commit %s", key.repo, key.ref, commit)
		c.cache.Store(key, &OrderResult{
			cancel: cancel,
			Commit: commit,
			Error:  err,
		})

		// cleanup cache entry after some time to avoid memory leaks
		<-timeoutCtx.Done()
		c.DeleteLastCommitCheckOrder(key.repo, key.ref, key.auth)
	}()
}

// IsLastCommitCheckOrdered checks if the last commit check has been ordered
func (c *asyncLastCommitChecker) IsLastCommitCheckOrdered(repo, ref string, auth *GitAuth) bool {
	key := orderCacheKey{
		repo: repo,
		ref:  ref,
		auth: auth,
	}
	_, exists := c.cache.Load(key)
	return exists
}

// DeleteLastCommitCheckOrder deletes the last commit check order
func (c *asyncLastCommitChecker) DeleteLastCommitCheckOrder(repo, ref string, auth *GitAuth) {
	key := orderCacheKey{
		repo: repo,
		ref:  ref,
		auth: auth,
	}

	result := c.load(key)
	if result == nil {
		return
	}

	result.cancel()
	c.cache.Delete(key)
	c.log.Debugf("cleaned up async last commit check cache for %s %s", key.repo, key.ref)
}

// GetLastCommitCheckResult gets the result of the last commit check
func (c *asyncLastCommitChecker) GetLastCommitCheckResult(repo, ref string, auth *GitAuth) *OrderResult {
	key := orderCacheKey{
		repo: repo,
		ref:  ref,
		auth: auth,
	}

	return c.load(key)
}

func (c *asyncLastCommitChecker) load(key orderCacheKey) *OrderResult {
	value, exists := c.cache.Load(key)
	if !exists {
		return nil
	}

	result, ok := value.(*OrderResult)
	if !ok {
		return nil
	}

	return result
}
