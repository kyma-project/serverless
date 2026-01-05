package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/cache"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func init() {
	// required by Azure Devops (works with Github, Gitlab, Bitbucket)
	// https://github.com/go-git/go-git/blob/master/_examples/azure_devops/main.go#L21-L36
	transport.UnsupportedCapabilities = []capability.Capability{
		capability.ThinPack,
	}
}

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
	repo, err := git.Init(memory.NewStorage(), nil)
	if err != nil {
		return "", err
	}

	remote, err := repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
	})
	if err != nil {
		return "", err
	}

	var auth transport.AuthMethod
	if gitAuth != nil {
		auth, err = gitAuth.GetAuthMethod()
		if err != nil {
			return "", errors.Wrap(err, "while choosing authorization method")
		}
	}

	refs, err := remote.List(&git.ListOptions{
		Auth: auth,
	})
	if err != nil {
		return "", err
	}

	for _, rf := range refs {
		rfName := rf.Name()
		if !rfName.IsBranch() && !rfName.IsTag() {
			continue
		}
		if rfName.Short() == reference {
			return rf.Hash().String(), nil
		}
	}

	return "", errors.New("reference not found")
}
