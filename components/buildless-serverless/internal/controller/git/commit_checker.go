package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
	crypto_ssh "golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
)

//go:generate mockery --name=LastCommitChecker --output=automock --outpkg=automock --case=underscore
type LastCommitChecker interface {
	GetLatestCommit(url, reference string, secret *corev1.Secret) (string, error)
}

type GoGitCommitChecker struct {
}

func (g GoGitCommitChecker) GetLatestCommit(url, reference string, secret *corev1.Secret) (string, error) {
	cloneOptions := git.CloneOptions{
		URL:           url,
		ReferenceName: plumbing.ReferenceName(reference),
		SingleBranch:  true,
		Depth:         1,
	}
	if secret != nil {
		auth, err := chooseAuth(secret)
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

func chooseAuth(secret *corev1.Secret) (transport.AuthMethod, error) {
	switch secret.Type {
	case "kubernetes.io/ssh-auth":
		return sshAuthForKubernetesSecret(secret)
	case "kubernetes.io/basic-auth":
		return basicAuthForKubernetesSecret(secret)
	default:
		// It is for compatibility with the previous implementation
		if _, keyFound := secret.Data["key"]; keyFound {
			return sshAuthForOldServerlessSecret(secret)
		}
		return basicAuthForOldServerlessSecret(secret)
	}
}

func basicAuthForOldServerlessSecret(secret *corev1.Secret) (transport.AuthMethod, error) {
	username, usernameFound := secret.Data["username"]
	password, passwordFound := secret.Data["password"]
	if !usernameFound || !passwordFound {
		return nil, errors.New("missing username, password or key")
	}
	return basicAuth(string(username), string(password))
}

func basicAuthForKubernetesSecret(secret *corev1.Secret) (transport.AuthMethod, error) {
	username, usernameFound := secret.Data["username"]
	password, passwordFound := secret.Data["password"]
	if !usernameFound || !passwordFound {
		return nil, errors.New("missing username or password")
	}
	return basicAuth(string(username), string(password))
}

func sshAuthForOldServerlessSecret(secret *corev1.Secret) (transport.AuthMethod, error) {
	key, keyFound := secret.Data["key"]
	if !keyFound {
		return nil, errors.New("missing key")
	}
	password, passwordFound := secret.Data["password"]
	if passwordFound {
		return sshAuth(key, string(password))
	}
	return sshAuth(key, "")
}

func sshAuthForKubernetesSecret(secret *corev1.Secret) (transport.AuthMethod, error) {
	privateKey, ok := secret.Data["ssh-privatekey"]
	if !ok {
		return nil, errors.New("missing ssh-privatekey")
	}
	return sshAuth(privateKey, "")
}

func sshAuth(sshPrivateKey []byte, sshPassword string) (transport.AuthMethod, error) {
	auth, err := ssh.NewPublicKeys("git", sshPrivateKey, sshPassword)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse private key")
	}

	// set callback to func that always returns nil while checking known hosts
	// this disables known hosts validation
	auth.HostKeyCallback = crypto_ssh.InsecureIgnoreHostKey()

	return auth, err
}

func basicAuth(username, password string) (transport.AuthMethod, error) {
	return &http.BasicAuth{
		Username: username,
		Password: password,
	}, nil
}
