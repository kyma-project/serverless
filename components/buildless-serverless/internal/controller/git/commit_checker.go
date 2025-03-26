package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/kyma-project/serverless/internal/controller/cache"
	"github.com/pkg/errors"
)

//go:generate mockery --name=LastCommitChecker --output=automock --outpkg=automock --case=underscore
type LastCommitChecker interface {
	GetLatestCommit(url, reference string, gitAuth *GitAuth, force bool) (string, error)
}

type GoGitCommitChecker struct {
	Cache cache.InMemoryCache
}

type LastCommitKey struct {
	url       string
	reference string
}

func (g GoGitCommitChecker) GetLatestCommit(url, reference string, gitAuth *GitAuth, force bool) (string, error) {
	commitKey := LastCommitKey{
		url:       url,
		reference: reference,
	}
	lastCommit := ""
	if !force {
		lastCommit = g.Cache.Get(commitKey)
	}
	if lastCommit != "" {
		return lastCommit, nil
	}

	cloneOptions := git.CloneOptions{
		URL:           url,
		ReferenceName: plumbing.ReferenceName(reference),
		SingleBranch:  true,
		Depth:         1,
	}
	if gitAuth != nil {
		auth, err := gitAuth.GetAuthMethod()
		if err != nil {
			return "", errors.Wrap(err, "while choosing authorization method")
		}
		cloneOptions.Auth = auth
	}

	r, err := git.Clone(memory.NewStorage(), nil, &cloneOptions)
	if err != nil {
		return "", err
	}

	ref, err := r.Head()
	if err != nil {
		return "", err
	}

	lastCommit = ref.Hash().String()
	g.Cache.Set(commitKey, lastCommit)

	return lastCommit, nil
}
