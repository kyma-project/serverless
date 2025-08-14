package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/cache"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//go:generate mockery --name=LastCommitChecker --output=automock --outpkg=automock --case=underscore
type LastCommitChecker interface {
	GetLatestCommit(url, reference string, gitAuth *GitAuth, force bool) (string, error)
}

type GoGitCachedCommitChecker struct {
	Cache cache.Cache
	Log   *zap.SugaredLogger
}

type LastCommitKey struct {
	url       string
	reference string
}

func (g GoGitCachedCommitChecker) GetLatestCommit(url, reference string, gitAuth *GitAuth, force bool) (string, error) {
	commitKey := LastCommitKey{
		url:       url,
		reference: reference,
	}
	if !force {
		cachedCommit := g.Cache.Get(commitKey)
		if cachedCommit != nil {
			g.Log.Debugf("Last commit from cache for %s %s is %s", url, reference, *cachedCommit)
			return *cachedCommit, nil
		}
	}

	lastCommit, err := GetLatestCommit(url, reference, gitAuth)
	if err != nil {
		return "", errors.Wrapf(err, "while getting latest commit for %s %s", url, reference)
	}

	g.Log.Debugf("Last commit from repository for %s %s is %s ", url, reference, lastCommit)
	g.Cache.Set(commitKey, lastCommit)

	return lastCommit, nil
}

func GetLatestCommit(url, reference string, gitAuth *GitAuth) (string, error) {
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

	// required by Azure Devops (works with Github, Gitlab, Bitbucket)
	// https://github.com/go-git/go-git/blob/master/_examples/azure_devops/main.go#L21-L36
	transport.UnsupportedCapabilities = []capability.Capability{
		capability.ThinPack,
	}

	r, err := git.Clone(memory.NewStorage(), nil, &cloneOptions)
	if err != nil {
		return "", err
	}

	ref, err := r.Head()
	if err != nil {
		return "", err
	}

	return ref.Hash().String(), nil
}
