package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	crypto_ssh "golang.org/x/crypto/ssh"
)

const envPrefix = "APP"

type initConfig struct {
	RepositoryURL       string
	RepositoryReference string
	RepositoryCommit    string
	DestinationPath     string
	//RepositoryAuthType git.RepositoryAuthType `envconfig:"optional"`
	RepositoryUsername string `envconfig:"optional"`
	RepositoryPassword string `envconfig:"optional"`
	RepositoryKey      string `envconfig:"optional"`
}

func main() {
	log.Println("Start repo fetcher...")

	cfg := initConfig{}
	if err := envconfig.InitWithPrefix(&cfg, envPrefix); err != nil {
		log.Fatalf("while reading env variables: %s", err.Error())
	}

	// test: list (in-memory) remotes
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{cfg.RepositoryURL},
	})

	fmt.Print(cfg.RepositoryKey)
	auth, err := ssh.NewPublicKeys("git", []byte(cfg.RepositoryKey), "")
	failOnErr(err)

	// set callback to func that always returns nil while checking known hosts
	// this disables known hosts validation
	auth.HostKeyCallback = crypto_ssh.InsecureIgnoreHostKey()

	fmt.Println("list remotes")
	rfs, _ := remote.List(&git.ListOptions{
		Auth: auth,
	})
	//failOnErr(err)

	fmt.Println("printing rfs")
	for _, rf := range rfs {
		fmt.Printf("Hash: %s\n\tName: %s\n\tType: %s\n\tTarget: %s\n", rf.Hash().String(), rf.Name().Short(), rf.Type().String(), rf.Target())
		fmt.Printf("\tIsTag: %v\n\tIsBranch: %v\n\tIsRemote: %v\n", rf.Name().IsTag(), rf.Name().IsBranch(), rf.Name().IsRemote())
	}

	//log.Println("Get auth config...")
	//gitOptions := cfg.getOptions()

	log.Printf("Clone repo from url: %s and commit: %s...\n", cfg.RepositoryURL, cfg.RepositoryCommit)
	err = clone(cfg)
	if err != nil {
		//if git.IsAuthErr(err) {
		//	log.Printf("while cloning repository bad credentials were provided, errMsg: %s", err.Error())
		//} else {
		log.Fatalln(errors.Wrapf(err, "while cloning repository: %s, from commit: %s", cfg.RepositoryURL, cfg.RepositoryCommit))
		//}
	}

	log.Printf("Cloned repository: %s, from commit: %s, to path: %s", cfg.RepositoryURL, cfg.RepositoryCommit, cfg.DestinationPath)
}

func clone(c initConfig) error {
	auth, _ := ssh.NewPublicKeys("git", []byte(c.RepositoryKey), "")
	//failOnErr(err)

	// set callback to func that always returns nil while checking known hosts
	// this disables known hosts validation
	auth.HostKeyCallback = crypto_ssh.InsecureIgnoreHostKey()

	r, err := git.PlainClone(c.DestinationPath, false, &git.CloneOptions{
		URL:           c.RepositoryURL,
		ReferenceName: plumbing.ReferenceName(c.RepositoryReference),
		SingleBranch:  true,
		// Auth: &http.BasicAuth{
		// 	Username: c.RepositoryUsername,
		// 	Password: c.RepositoryPassword,
		// },
		Auth: auth,
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

func failOnErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
