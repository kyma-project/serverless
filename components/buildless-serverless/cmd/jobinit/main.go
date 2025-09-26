package main

import (
	"crypto/fips140"
	"fmt"
	"log"
	"os"

	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/vrischmann/envconfig"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/pkg/errors"
	crypto_ssh "golang.org/x/crypto/ssh"
)

const envPrefix = "APP"

type initConfig struct {
	RepositoryURL       string
	RepositoryReference string
	RepositoryCommit    string
	DestinationPath     string
	RepositoryAuthType  serverlessv1alpha2.RepositoryAuthType `envconfig:"optional"`
	RepositoryUsername  string                                `envconfig:"optional"`
	RepositoryPassword  string                                `envconfig:"optional"`
	RepositoryKey       string                                `envconfig:"optional"`
}

func main() {
	log.Println("Start repo fetcher...")

	if !isFIPS140Only() {
		log.Panic("FIPS 140 exclusive mode is not enabled. Check GODEBUG flags.")
	}

	cfg := initConfig{}
	if err := envconfig.InitWithPrefix(&cfg, envPrefix); err != nil {
		log.Fatalf("while reading env variables: %s", err.Error())
	}

	auth, err := chooseAuth(cfg)
	failOnErr(err, "unable to choose auth")

	// required by Azure Devops (works with Github, Gitlab, Bitbucket)
	// https://github.com/go-git/go-git/blob/master/_examples/azure_devops/main.go#L21-L36
	transport.UnsupportedCapabilities = []capability.Capability{
		capability.ThinPack,
	}

	log.Printf("Clone repo from url: %s and commit: %s...\n", cfg.RepositoryURL, cfg.RepositoryCommit)
	err = clone(cfg, auth)
	failOnErr(err, "while cloning repository from commit")

	log.Printf("Cloned repository: %s, from commit: %s, to path: %s", cfg.RepositoryURL, cfg.RepositoryCommit, cfg.DestinationPath)
}

func clone(c initConfig, auth transport.AuthMethod) error {
	r, err := git.PlainClone(c.DestinationPath, false, &git.CloneOptions{
		URL:           c.RepositoryURL,
		ReferenceName: plumbing.ReferenceName(c.RepositoryReference),
		SingleBranch:  true,
		Auth:          auth,
	})
	if err != nil {
		return err
	}

	wt, err := r.Worktree()
	if err != nil {
		return err
	}

	err = wt.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(c.RepositoryCommit),
	})
	if err != nil {
		return err
	}

	return nil
}

func failOnErr(err error, msg string) {
	if err != nil {
		if msg != "" {
			err = errors.Wrap(err, msg)
		}
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func chooseAuth(cfg initConfig) (transport.AuthMethod, error) {
	if cfg.RepositoryAuthType == "" {
		log.Printf("Repository auth type not provided, skipping authorization")
		return nil, nil
	}
	switch cfg.RepositoryAuthType {
	case serverlessv1alpha2.RepositoryAuthSSHKey:
		return sshAuth([]byte(cfg.RepositoryKey), cfg.RepositoryPassword)
	case serverlessv1alpha2.RepositoryAuthBasic:
		return basicAuth(cfg.RepositoryUsername, cfg.RepositoryPassword)
	default:
		return nil, fmt.Errorf("unknown repository auth type: %s", cfg.RepositoryAuthType)
	}
}

func sshAuth(sshPrivateKey []byte, sshPassword string) (transport.AuthMethod, error) {
	auth, err := ssh.NewPublicKeys("git", sshPrivateKey, sshPassword)
	failOnErr(err, "unable to parse private key")

	// set callback to func that always returns nil while checking known hosts
	// this disables known hosts validation
	auth.HostKeyCallback = crypto_ssh.InsecureIgnoreHostKey()

	return auth, nil
}

func basicAuth(username, password string) (transport.AuthMethod, error) {
	return &http.BasicAuth{
		Username: username,
		Password: password,
	}, nil
}

func isFIPS140Only() bool {
	return fips140.Enabled() && os.Getenv("GODEBUG") == "fips140=only"
}
