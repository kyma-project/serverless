package git

import (
	"fmt"
	"net"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

const (
	UsernameKey = "username"
	PasswordKey = "password"
	KeyKey      = "key"
)

type RepositoryAuthType string

const (
	RepositoryAuthBasic  RepositoryAuthType = "basic"
	RepositoryAuthSSHKey RepositoryAuthType = "key"
)

type AuthOptions struct {
	Type        RepositoryAuthType
	Credentials map[string]string
	SecretName  string
}

func GetGoGitAuth(options *AuthOptions) (transport.AuthMethod, error) {
	if options == nil {
		return nil, nil
	}

	switch authType := options.Type; authType {
	case RepositoryAuthBasic:
		basic, err := toGoGitBasicAuth(options)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting authentication config to %s auth method", authType)
		}
		return transport.AuthMethod(basic), nil
	case RepositoryAuthSSHKey:
		key, err := toGoGitKeyAuth(options)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting authentication config to %s auth method", authType)
		}
		return transport.AuthMethod(key), nil
	default:
		return nil, fmt.Errorf("unknown authentication type: %s", authType)
	}
}

func toGoGitBasicAuth(options *AuthOptions) (*http.BasicAuth, error) {
	if options.Credentials == nil {
		return &http.BasicAuth{}, nil
	}

	username, ok := options.Credentials[UsernameKey]
	if !ok {
		return nil, fmt.Errorf("missing field %s", UsernameKey)
	}

	password, ok := options.Credentials[PasswordKey]
	if !ok {
		return nil, fmt.Errorf("missing field %s", PasswordKey)
	}

	return &http.BasicAuth{
		Username: username,
		Password: password,
	}, nil
}

func toGoGitKeyAuth(options *AuthOptions) (*gitssh.PublicKeys, error) {
	if options.Credentials == nil {
		return &gitssh.PublicKeys{}, nil
	}

	key, ok := options.Credentials[KeyKey]
	if !ok {
		return nil, fmt.Errorf("missing field %s", KeyKey)
	}

	password, _ := options.Credentials[PasswordKey]

	var signer ssh.Signer
	var err error
	if password == "" {
		signer, err = ssh.ParsePrivateKey([]byte(key))
	} else {
		signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(key), []byte(password))
	}
	if err != nil {
		return nil, errors.Wrapf(err, "while creating public keys authentication method")
	}

	auth := gitssh.PublicKeys{
		User:   "git",
		Signer: signer,
		HostKeyCallbackHelper: gitssh.HostKeyCallbackHelper{HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		}},
	}

	return &auth, nil
}
