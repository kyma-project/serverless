package main

import (
	"log"

	"github.com/kyma-project/serverless/components/serverless/internal/git"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
)

const envPrefix = "APP"

type config struct {
	RepositoryURL      string
	RepositoryCommit   string
	MountPath          string                 `envconfig:"default=/workspace"`
	RepositoryAuthType git.RepositoryAuthType `envconfig:"optional"`
	RepositoryUsername string                 `envconfig:"optional"`
	RepositoryPassword string                 `envconfig:"optional"`
	RepositoryKey      string                 `envconfig:"optional"`
}

func main() {
	log.Println("Start repo fetcher...")
	cfg := config{}
	if err := envconfig.InitWithPrefix(&cfg, envPrefix); err != nil {
		log.Fatalf("while reading env variables: %s", err.Error())
	}

	logger, _ := zap.NewProduction()
	operator := git.NewMixedClient(logger.Sugar())

	log.Println("Get auth config...")
	gitOptions := cfg.getOptions()

	log.Printf("Clone repo from url: %s and commit: %s...\n", cfg.RepositoryURL, cfg.RepositoryCommit)
	commit, err := operator.Clone(cfg.MountPath, gitOptions)
	if err != nil {
		log.Fatalln(errors.Wrapf(err, "while cloning repository: %s from commit: %s: %s", cfg.RepositoryURL, cfg.RepositoryCommit, err.Error()))
	}

	log.Printf("Cloned repository: %s, from commit: %s, to path: %s", cfg.RepositoryURL, commit, cfg.MountPath)
}

func (c *config) getOptions() git.Options {
	return git.Options{
		URL:       c.RepositoryURL,
		Reference: c.RepositoryCommit,
		Auth:      c.getAuthFromType(),
	}
}

func (c *config) getAuthFromType() *git.AuthOptions {
	switch c.RepositoryAuthType {
	case git.RepositoryAuthBasic:
		return &git.AuthOptions{
			Type: git.RepositoryAuthBasic,
			Credentials: map[string]string{
				git.UsernameKey: c.RepositoryUsername,
				git.PasswordKey: c.RepositoryPassword,
			},
		}
	case git.RepositoryAuthSSHKey:
		return &git.AuthOptions{
			Type: git.RepositoryAuthSSHKey,
			Credentials: map[string]string{
				git.KeyKey:      c.RepositoryKey,
				git.PasswordKey: c.RepositoryPassword,
			},
		}
	default:
		return nil
	}
}
