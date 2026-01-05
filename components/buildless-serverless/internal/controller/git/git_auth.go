package git

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/pkg/errors"
	crypto_ssh "golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type dataField[T any] struct {
	value     T
	fieldName string
	envName   string
}

type GitAuth struct {
	secretName      string
	secretNamespace string
	client          client.Client
	secret          *corev1.Secret
	authType        serverlessv1alpha2.RepositoryAuthType
	username        *dataField[string]
	password        *dataField[string]
	sshKey          *dataField[[]byte]
}

func NewGitAuth(ctx context.Context, client client.Client, f *serverlessv1alpha2.Function) (*GitAuth, error) {
	a := &GitAuth{
		secretName:      f.Spec.Source.GitRepository.Auth.SecretName,
		secretNamespace: f.GetNamespace(),
		authType:        f.Spec.Source.GitRepository.Auth.Type,
		client:          client,
	}
	err := a.loadSecret(ctx)
	if err != nil {
		return nil, err
	}
	err = a.parseSecret()
	if err != nil {
		return nil, errors.Wrap(err, "while parsing git authorization secret")
	}
	return a, nil
}

func (a *GitAuth) loadSecret(ctx context.Context) error {
	s := &corev1.Secret{}
	err := a.client.Get(ctx,
		types.NamespacedName{
			Namespace: a.secretNamespace,
			Name:      a.secretName,
		}, s)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to get secret: %s", err.Error()))
	}
	a.secret = s
	return nil
}

func (a *GitAuth) parseSecret() error {
	switch a.secret.Type {
	case corev1.SecretTypeSSHAuth:
		return a.parseSSHAuthKubernetesSecret()
	case corev1.SecretTypeBasicAuth:
		return a.parseBasicAuthKubernetesSecret()
		// It is for compatibility with the previous implementation
	default:
		switch a.authType {
		case serverlessv1alpha2.RepositoryAuthSSHKey:
			return a.parseSSHAuthOldServerlessSecret()
		case serverlessv1alpha2.RepositoryAuthBasic:
			return a.parseBasicAuthOldServerlessSecret()
		default:
			return errors.New("unexpected authorization type")
		}
	}
}

func (a *GitAuth) GetAuthMethod() (transport.AuthMethod, error) {
	switch a.authType {
	case serverlessv1alpha2.RepositoryAuthSSHKey:
		return a.sshAuth()
	case serverlessv1alpha2.RepositoryAuthBasic:
		return a.basicAuth()
	default:
		return nil, errors.New("unexpected authorization type")
	}
}

func (a *GitAuth) GetAuthEnvs() []corev1.EnvVar {
	s := a.secretName
	var envs []corev1.EnvVar
	envs = append(envs, corev1.EnvVar{
		Name:  repositoryAuthTypeEnvVarName,
		Value: string(a.authType),
	})
	envs = addEnvVar(envs, a.sshKey, s)
	envs = addEnvVar(envs, a.username, s)
	envs = addEnvVar(envs, a.password, s)
	return envs
}

const (
	kubernetesKeyFieldName         = "ssh-privatekey"
	kubernetesUsernameFieldName    = "username"
	kubernetesPasswordFieldName    = "password"
	oldServerlessKeyFieldName      = "key"
	oldServerlessUsernameFieldName = "username"
	oldServerlessPasswordFieldName = "password"
	repositoryAuthTypeEnvVarName   = "APP_REPOSITORY_AUTH_TYPE"
	usernameEnvVarName             = "APP_REPOSITORY_USERNAME"
	passwordEnvVarName             = "APP_REPOSITORY_PASSWORD"
	sshKeyEnvVarName               = "APP_REPOSITORY_KEY"
)

func (a *GitAuth) parseSSHAuthKubernetesSecret() error {
	if a.authType != serverlessv1alpha2.RepositoryAuthSSHKey {
		return errors.New(fmt.Sprintf("inconsistent secret type: %s, %s", a.authType, corev1.SecretTypeSSHAuth))
	}
	privateKey, ok := a.secret.Data[kubernetesKeyFieldName]
	if !ok {
		return errors.New(fmt.Sprintf("missing '%s'", kubernetesKeyFieldName))
	}
	a.sshKey = &dataField[[]byte]{
		value:     privateKey,
		fieldName: kubernetesKeyFieldName,
		envName:   sshKeyEnvVarName,
	}
	return nil
}

func (a *GitAuth) parseBasicAuthKubernetesSecret() error {
	if a.authType != serverlessv1alpha2.RepositoryAuthBasic {
		return errors.New(fmt.Sprintf("inconsistent secret type: %s, %s", a.authType, corev1.SecretTypeBasicAuth))
	}
	username, usernameFound := a.secret.Data[kubernetesUsernameFieldName]
	password, passwordFound := a.secret.Data[kubernetesPasswordFieldName]
	if !usernameFound || !passwordFound {
		return errors.New(fmt.Sprintf("missing '%s' or '%s'", kubernetesUsernameFieldName, kubernetesPasswordFieldName))
	}
	a.username = &dataField[string]{
		value:     string(username),
		fieldName: kubernetesUsernameFieldName,
		envName:   usernameEnvVarName,
	}
	a.password = &dataField[string]{
		value:     string(password),
		fieldName: kubernetesPasswordFieldName,
		envName:   passwordEnvVarName,
	}
	return nil
}

func (a *GitAuth) parseSSHAuthOldServerlessSecret() error {
	key, keyFound := a.secret.Data[oldServerlessKeyFieldName]
	if !keyFound {
		return errors.New(fmt.Sprintf("missing '%s'", oldServerlessKeyFieldName))
	}
	a.sshKey = &dataField[[]byte]{
		value:     key,
		fieldName: oldServerlessKeyFieldName,
		envName:   sshKeyEnvVarName,
	}
	password, passwordFound := a.secret.Data[oldServerlessPasswordFieldName]
	if passwordFound {
		a.password = &dataField[string]{
			value:     string(password),
			fieldName: oldServerlessPasswordFieldName,
			envName:   passwordEnvVarName,
		}
	}
	return nil
}

func (a *GitAuth) parseBasicAuthOldServerlessSecret() error {
	username, usernameFound := a.secret.Data[oldServerlessUsernameFieldName]
	password, passwordFound := a.secret.Data[oldServerlessPasswordFieldName]
	if !usernameFound || !passwordFound {
		return errors.New(fmt.Sprintf("missing '%s' or '%s'", oldServerlessUsernameFieldName, oldServerlessPasswordFieldName))
	}
	a.username = &dataField[string]{
		value:     string(username),
		fieldName: oldServerlessUsernameFieldName,
		envName:   usernameEnvVarName,
	}
	a.password = &dataField[string]{
		value:     string(password),
		fieldName: oldServerlessPasswordFieldName,
		envName:   passwordEnvVarName,
	}
	return nil
}

func (a *GitAuth) sshAuth() (transport.AuthMethod, error) {
	password := ""
	if a.password != nil {
		password = a.password.value
	}
	auth, err := ssh.NewPublicKeys("git", a.sshKey.value, password)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse private key")
	}

	// set callback to func that always returns nil while checking known hosts
	// this disables known hosts validation
	auth.HostKeyCallback = crypto_ssh.InsecureIgnoreHostKey()

	return auth, nil
}

func (a *GitAuth) basicAuth() (transport.AuthMethod, error) {
	return &http.BasicAuth{
		Username: a.username.value,
		Password: a.password.value,
	}, nil
}

func addEnvVar[T any](envs []corev1.EnvVar, f *dataField[T], secretName string) []corev1.EnvVar {
	if f == nil {
		return envs
	}
	envs = append(envs, corev1.EnvVar{
		Name: f.envName,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
				Key: f.fieldName,
			},
		},
	})
	return envs
}
