package main

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"log"
)

const envPrefix = "APP"

type config struct {
	RepositoryURL       string
	RepositoryReference string
	RepositoryCommit    string
	DestinationPath     string
	//RepositoryAuthType git.RepositoryAuthType `envconfig:"optional"`
	//RepositoryUsername string                 `envconfig:"optional"`
	//RepositoryPassword string                 `envconfig:"optional"`
	//RepositoryKey      string                 `envconfig:"optional"`
}

func main() {
	log.Println("Start repo fetcher...")

	cfg := config{}
	if err := envconfig.InitWithPrefix(&cfg, envPrefix); err != nil {
		log.Fatalf("while reading env variables: %s", err.Error())
	}

	//log.Println("Get auth config...")
	//gitOptions := cfg.getOptions()

	log.Printf("Clone repo from url: %s and commit: %s...\n", cfg.RepositoryURL, cfg.RepositoryCommit)
	err := clone(cfg)
	if err != nil {
		//if git.IsAuthErr(err) {
		//	log.Printf("while cloning repository bad credentials were provided, errMsg: %s", err.Error())
		//} else {
		log.Fatalln(errors.Wrapf(err, "while cloning repository: %s, from commit: %s", cfg.RepositoryURL, cfg.RepositoryCommit))
		//}
	}

	log.Printf("Cloned repository: %s, from commit: %s, to path: %s", cfg.RepositoryURL, cfg.RepositoryCommit, cfg.DestinationPath)
}

func clone(c config) error {
	r, err := git.PlainClone(c.DestinationPath, false, &git.CloneOptions{
		URL:           c.RepositoryURL,
		ReferenceName: plumbing.ReferenceName(c.RepositoryReference),
		SingleBranch:  true,
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
