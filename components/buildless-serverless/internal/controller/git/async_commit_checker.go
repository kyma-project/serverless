package git

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

//go:generate mockery --name=AsyncLatestCommitChecker --output=automock --outpkg=automock --case=underscore
type AsyncLatestCommitChecker interface {
	PlaceOrder(string, string, string, *GitAuth)
	CollectOrder(string) *OrderResult
}

type asyncLatestCommitChecker struct {
	ctx               context.Context
	cache             sync.Map
	log               *zap.SugaredLogger
	cacheElemLifetime time.Duration

	// implemented to allow easier testing
	getLatestCommit func(repo, ref string, auth *GitAuth) (string, error)
}

type OrderResult struct {
	Commit    string
	Error     error
	timestamp time.Time
}

func NewAsyncLatestCommitChecker(ctx context.Context, log *zap.SugaredLogger) AsyncLatestCommitChecker {
	checker := &asyncLatestCommitChecker{
		ctx:               ctx,
		log:               log,
		getLatestCommit:   GetLatestCommit,
		cacheElemLifetime: 2 * time.Minute,
	}

	// start periodic cache cleanup
	checker.clearCacheEvery(time.Hour * 24)

	return checker
}

// PlaceOrder orders asynchronous git latest commit check
// when the check is complete, the result can be accessed using the orderID
func (c *asyncLatestCommitChecker) PlaceOrder(orderID, repo, ref string, auth *GitAuth) {
	_, exists := c.cache.Load(orderID)
	if exists {
		// already ordered
		return
	}

	// store a nil value to indicate that the order is in progress
	c.cache.Store(orderID, nil)

	go func() {
		c.log.Debugf("starting async latest commit check for %s %s", repo, ref)
		commit, err := c.getLatestCommit(repo, ref, auth)

		c.log.Debugf("finished async lalatestst commit check for %s %s with commit %s", repo, ref, commit)
		c.cache.Store(orderID, &OrderResult{
			Commit:    commit,
			Error:     err,
			timestamp: time.Now(),
		})
	}()
}

// CollectOrder collects the result of the latest commit check for the given orderID
// if the result is found or the order is still in progress, nil is returned
// if order is older than 2 minutes, it is removed from the cache but latest order is returned
func (c *asyncLatestCommitChecker) CollectOrder(orderID string) *OrderResult {
	result := c.load(orderID)
	if result != nil && time.Since(result.timestamp) > c.cacheElemLifetime {
		// remove old result from cache if is older than 2 minutes
		c.cache.Delete(orderID)
	}

	return result
}

func (c *asyncLatestCommitChecker) load(orderID string) *OrderResult {
	value, exists := c.cache.Load(orderID)
	if !exists {
		return nil
	}

	result, ok := value.(*OrderResult)
	if !ok {
		return nil
	}

	return result
}

func (c *asyncLatestCommitChecker) clearCacheEvery(duration time.Duration) {
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			case <-time.After(duration):
				c.log.Debug("clearing async latest commit checker cache")
				c.cache.Clear()
			}
		}
	}()
}
