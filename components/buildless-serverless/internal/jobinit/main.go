package main

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/vrischmann/envconfig"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"

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
	AuthSecretName      string `envconfig:"optional"`
}

func main() {
	log.Println("Start repo fetcher...")

	cfg := initConfig{}
	if err := envconfig.InitWithPrefix(&cfg, envPrefix); err != nil {
		log.Fatalf("while reading env variables: %s", err.Error())
	}

	restConfig, err := restConfig("")
	failOnErr(err, "unable to load k8s config")

	client, err := kubernetes.NewForConfig(restConfig)
	failOnErr(err, "unable to create a client")

	secret, err := client.CoreV1().Secrets("default").Get(context.Background(), cfg.AuthSecretName, metav1.GetOptions{})
	failOnErr(err, "unable to get secret")

	auth, err := chooseAuth(secret)
	failOnErr(err, "unable to choose auth")

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

// restConfig loads the rest configuration needed by k8s clients to interact with clusters based on the kubeconfig.
// Loading rules are based on standard defined kubernetes config loading.
func restConfig(kubeconfig string) (*rest.Config, error) {
	// Default PathOptions gets kubeconfig in this order: the explicit path given, KUBECONFIG current context, recommended file path
	po := clientcmd.NewDefaultPathOptions()
	po.LoadingRules.ExplicitPath = kubeconfig

	cfg, err := clientcmd.BuildConfigFromKubeconfigGetter("", po.GetStartingConfig)
	if err != nil {
		return nil, err
	}
	cfg.WarningHandler = rest.NoWarnings{}
	return cfg, nil
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
