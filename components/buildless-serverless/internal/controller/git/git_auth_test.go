package git

import (
	"context"
	"encoding/json"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestGitAuth_LoadSecret(t *testing.T) {
	t.Run("successfully loaded secret", func(t *testing.T) {
		// Arrange
		// some secret on k8s
		someSecret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gracious-robinson",
				Namespace: "suspicious-bhabha"},
			Data: map[string][]byte{
				"turing": []byte("bold"),
			}}

		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, corev1.AddToScheme(scheme))
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&someSecret).Build()
		// git authorization with data to load
		a := &GitAuth{
			secretName:      "gracious-robinson",
			secretNamespace: "suspicious-bhabha",
			client:          k8sClient,
		}

		// Act
		err := a.loadSecret(context.Background())

		// Assert
		require.NoError(t, err)
		require.NotNil(t, a.secret)
		require.Equal(t, "bold", string(someSecret.Data["turing"]))
	})
	t.Run("failed to load secret", func(t *testing.T) {
		// Arrange
		// scheme and fake client without secrets
		scheme := runtime.NewScheme()
		require.NoError(t, corev1.AddToScheme(scheme))
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		// git authorization with data to load
		a := &GitAuth{
			secretName:      "mystifying-khorana",
			secretNamespace: "reverent-saha",
			client:          k8sClient,
		}

		// Act
		err := a.loadSecret(context.Background())

		// Assert
		require.Error(t, err)
		require.EqualError(t, err, "failed to get secret: secrets \"mystifying-khorana\" not found")
		require.Nil(t, a.secret)
	})
}

func TestGitAuth_ParseSecret(t *testing.T) {
	type fields struct {
		secret   *corev1.Secret
		authType serverlessv1alpha2.RepositoryAuthType
	}
	type want struct {
		isError      bool
		errorMessage string
		username     *dataField[string]
		password     *dataField[string]
		sshKey       *dataField[[]byte]
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "inconsistent kubernetes secret with SSH key vs auth type basic",
			fields: fields{
				secret: &corev1.Secret{
					Type: corev1.SecretTypeSSHAuth,
				},
				authType: serverlessv1alpha2.RepositoryAuthBasic,
			},
			want: want{
				isError:      true,
				errorMessage: "inconsistent secret type: basic, kubernetes.io/ssh-auth",
			},
		},
		{
			name: "missing data in kubernetes secret with SSH key",
			fields: fields{
				secret: &corev1.Secret{
					Type: corev1.SecretTypeSSHAuth,
				},
				authType: serverlessv1alpha2.RepositoryAuthSSHKey,
			},
			want: want{
				isError:      true,
				errorMessage: "missing 'ssh-privatekey'",
			},
		},
		{
			name: "proper kubernetes secret with SSH key",
			fields: fields{
				secret: &corev1.Secret{
					Type: corev1.SecretTypeSSHAuth,
					Data: map[string][]byte{
						"ssh-privatekey": []byte("vigilant-buck"),
					},
				},
				authType: serverlessv1alpha2.RepositoryAuthSSHKey,
			},
			want: want{
				isError: false,
				sshKey: &dataField[[]byte]{
					value:     []byte("vigilant-buck"),
					fieldName: "ssh-privatekey",
					envName:   sshKeyEnvVarName,
				},
			},
		},
		{
			name: "inconsistent kubernetes secret with basic auth vs auth type SSH key",
			fields: fields{
				secret: &corev1.Secret{
					Type: corev1.SecretTypeBasicAuth,
				},
				authType: serverlessv1alpha2.RepositoryAuthSSHKey,
			},
			want: want{
				isError:      true,
				errorMessage: "inconsistent secret type: key, kubernetes.io/basic-auth",
			},
		},
		{
			name: "missing password in kubernetes secret with basic auth",
			fields: fields{
				secret: &corev1.Secret{
					Type: corev1.SecretTypeBasicAuth,
					Data: map[string][]byte{
						"username": []byte("interesting-volhard"),
					},
				},
				authType: serverlessv1alpha2.RepositoryAuthBasic,
			},
			want: want{
				isError:      true,
				errorMessage: "missing 'username' or 'password'",
			},
		},
		{
			name: "missing username in kubernetes secret with basic auth",
			fields: fields{
				secret: &corev1.Secret{
					Type: corev1.SecretTypeBasicAuth,
					Data: map[string][]byte{
						"password": []byte("agitated-morse"),
					},
				},
				authType: serverlessv1alpha2.RepositoryAuthBasic,
			},
			want: want{
				isError:      true,
				errorMessage: "missing 'username' or 'password'",
			},
		},
		{
			name: "proper kubernetes secret with basic auth",
			fields: fields{
				secret: &corev1.Secret{
					Type: corev1.SecretTypeBasicAuth,
					Data: map[string][]byte{
						"username": []byte("pedantic-jackson"),
						"password": []byte("intelligent-ardinghelli"),
					},
				},
				authType: serverlessv1alpha2.RepositoryAuthBasic,
			},
			want: want{
				isError: false,
				username: &dataField[string]{
					value:     "pedantic-jackson",
					fieldName: "username",
					envName:   usernameEnvVarName,
				},
				password: &dataField[string]{
					value:     "intelligent-ardinghelli",
					fieldName: "password",
					envName:   passwordEnvVarName,
				},
			},
		},
		{
			name: "missing data in old serverless secret with SSH key",
			fields: fields{
				secret: &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
				},
				authType: serverlessv1alpha2.RepositoryAuthSSHKey,
			},
			want: want{
				isError:      true,
				errorMessage: "missing 'key'",
			},
		},
		{
			name: "proper old serverless secret with SSH key",
			fields: fields{
				secret: &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"key": []byte("youthful-wescoff"),
					},
				},
				authType: serverlessv1alpha2.RepositoryAuthSSHKey,
			},
			want: want{
				isError: false,
				sshKey: &dataField[[]byte]{
					value:     []byte("youthful-wescoff"),
					fieldName: "key",
					envName:   sshKeyEnvVarName,
				},
			},
		},
		{
			name: "proper old serverless secret with SSH key and password",
			fields: fields{
				secret: &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"key":      []byte("vigorous-chaplygin"),
						"password": []byte("hungry-herschel"),
					},
				},
				authType: serverlessv1alpha2.RepositoryAuthSSHKey,
			},
			want: want{
				isError: false,
				password: &dataField[string]{
					value:     "hungry-herschel",
					fieldName: "password",
					envName:   passwordEnvVarName,
				},
				sshKey: &dataField[[]byte]{
					value:     []byte("vigorous-chaplygin"),
					fieldName: "key",
					envName:   sshKeyEnvVarName,
				},
			},
		},

		{
			name: "missing password in old serverless secret with basic auth",
			fields: fields{
				secret: &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"username": []byte("clever-brahmagupta"),
					},
				},
				authType: serverlessv1alpha2.RepositoryAuthBasic,
			},
			want: want{
				isError:      true,
				errorMessage: "missing 'username' or 'password'",
			},
		},
		{
			name: "missing username in old serverless secret with basic auth",
			fields: fields{
				secret: &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"password": []byte("eloquent-joliot"),
					},
				},
				authType: serverlessv1alpha2.RepositoryAuthBasic,
			},
			want: want{
				isError:      true,
				errorMessage: "missing 'username' or 'password'",
			},
		},
		{
			name: "proper old serverless secret with basic auth",
			fields: fields{
				secret: &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"username": []byte("zealous-moore"),
						"password": []byte("gifted-rhodes"),
					},
				},
				authType: serverlessv1alpha2.RepositoryAuthBasic,
			},
			want: want{
				isError: false,
				username: &dataField[string]{
					value:     "zealous-moore",
					fieldName: "username",
					envName:   usernameEnvVarName,
				},
				password: &dataField[string]{
					value:     "gifted-rhodes",
					fieldName: "password",
					envName:   passwordEnvVarName,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			a := &GitAuth{
				secret:   tt.fields.secret,
				authType: tt.fields.authType,
			}
			// Act
			err := a.parseSecret()
			// Assert
			if tt.want.isError {
				require.Error(t, err)
				require.EqualError(t, err, tt.want.errorMessage)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want.username, a.username)
				require.Equal(t, tt.want.password, a.password)
				require.Equal(t, tt.want.sshKey, a.sshKey)
			}
		})
	}
}

func TestGitAuth_GetAuthMethod(t *testing.T) {
	type fields struct {
		authType serverlessv1alpha2.RepositoryAuthType
		username *dataField[string]
		password *dataField[string]
		sshKey   *dataField[[]byte]
	}
	type want struct {
		isError          bool
		errorMessage     string
		authMethodString string
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "basic auth",
			fields: fields{
				authType: serverlessv1alpha2.RepositoryAuthBasic,
				username: &dataField[string]{
					value: "elegant-brown",
				},
				password: &dataField[string]{
					value: "strange-jennings",
				},
			},
			want: want{
				isError:          false,
				authMethodString: "http-basic-auth - elegant-brown:*******",
			},
		},
		{
			name: "ssh auth - invalid key",
			fields: fields{
				authType: serverlessv1alpha2.RepositoryAuthSSHKey,
				sshKey: &dataField[[]byte]{
					value: []byte("hardcore-ishizaka"),
				},
			},
			want: want{
				isError:      true,
				errorMessage: "unable to parse private key: ssh: no key found",
			},
		},
		// TODO: add tests for proper ssh key (and password)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &GitAuth{
				authType: tt.fields.authType,
				username: tt.fields.username,
				password: tt.fields.password,
				sshKey:   tt.fields.sshKey,
			}
			r, err := a.GetAuthMethod()
			if tt.want.isError {
				require.Error(t, err)
				require.EqualError(t, err, tt.want.errorMessage)
				require.Nil(t, r)
			} else {
				require.NoError(t, err)
				require.NotNil(t, r)
				require.Equal(t, tt.want.authMethodString, r.String())
			}
		})
	}
}

func TestGitAuth_GetAuthEnvs(t *testing.T) {
	type fields struct {
		secretName string
		authType   serverlessv1alpha2.RepositoryAuthType
		username   *dataField[string]
		password   *dataField[string]
		sshKey     *dataField[[]byte]
	}
	tests := []struct {
		name   string
		fields fields
		want   []corev1.EnvVar
	}{
		{
			name: "ssh key",
			fields: fields{
				authType:   serverlessv1alpha2.RepositoryAuthSSHKey,
				secretName: "quizzical-goodall",
				sshKey: &dataField[[]byte]{
					envName:   "clever-meninsky",
					fieldName: "laughing-dhawan",
				},
			},
			want: []corev1.EnvVar{
				{
					Name:  repositoryAuthTypeEnvVarName,
					Value: string(serverlessv1alpha2.RepositoryAuthSSHKey),
				},
				{
					Name:      "clever-meninsky",
					ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "quizzical-goodall"}, Key: "laughing-dhawan"}},
				},
			},
		},
		{
			name: "basic auth",
			fields: fields{
				authType:   serverlessv1alpha2.RepositoryAuthBasic,
				secretName: "crazy-easley",
				username: &dataField[string]{
					envName:   "friendly-keller",
					fieldName: "reverent-allen",
				},
				password: &dataField[string]{
					envName:   "naughty-mayer",
					fieldName: "exciting-lederberg",
				},
			},
			want: []corev1.EnvVar{
				{
					Name:  repositoryAuthTypeEnvVarName,
					Value: string(serverlessv1alpha2.RepositoryAuthBasic),
				},
				{
					Name:      "friendly-keller",
					ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "crazy-easley"}, Key: "reverent-allen"}},
				},
				{
					Name:      "naughty-mayer",
					ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "crazy-easley"}, Key: "exciting-lederberg"}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			a := &GitAuth{
				authType:   tt.fields.authType,
				secretName: tt.fields.secretName,
				username:   tt.fields.username,
				password:   tt.fields.password,
				sshKey:     tt.fields.sshKey,
			}
			// Act
			r := a.GetAuthEnvs()
			// Assert
			require.NotNil(t, r)
			jsonWant, _ := json.Marshal(tt.want)
			jsonR, _ := json.Marshal(r)
			require.Equal(t, string(jsonWant), string(jsonR))
		})
	}
}
