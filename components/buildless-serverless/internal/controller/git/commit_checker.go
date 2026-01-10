package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
)

func init() {
	// required by Azure Devops (works with Github, Gitlab, Bitbucket)
	// https://github.com/go-git/go-git/blob/master/_examples/azure_devops/main.go#L21-L36
	transport.UnsupportedCapabilities = []capability.Capability{
		capability.ThinPack,
	}
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
