package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
)

//go:generate mockery --name=LastCommitChecker --output=automock --outpkg=automock --case=underscore
type LastCommitChecker interface {
	GetLatestCommit(url, reference string, gitAuth *GitAuth) (string, error)
}

type GoGitCommitChecker struct {
}

func (g GoGitCommitChecker) GetLatestCommit(url, reference string, gitAuth *GitAuth) (string, error) {
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

	return ref.Hash().String(), nil
}
