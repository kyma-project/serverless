package serverless

import (
	"github.com/kyma-project/serverless/components/serverless/internal/git"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	destinationArg        = "--destination"
	functionContainerName = "function"
	baseDir               = "/workspace/src/"
	workspaceMountPath    = "/workspace"
)

var (
	istioSidecarInjectFalse = map[string]string{
		"sidecar.istio.io/inject": "false",
	}
	svcTargetPort = intstr.FromInt(8080)
)

func buildRepoFetcherEnvVars(instance *serverlessv1alpha2.Function, gitOptions git.Options) []corev1.EnvVar {
	vars := []corev1.EnvVar{
		{
			Name:  "APP_REPOSITORY_URL",
			Value: gitOptions.URL,
		},
		{
			Name:  "APP_REPOSITORY_COMMIT",
			Value: instance.Status.Commit,
		},
		{
			Name:  "APP_MOUNT_PATH",
			Value: workspaceMountPath,
		},
	}

	if gitOptions.Auth != nil {
		vars = append(vars, corev1.EnvVar{
			Name:  "APP_REPOSITORY_AUTH_TYPE",
			Value: string(gitOptions.Auth.Type),
		})

		switch gitOptions.Auth.Type {
		case git.RepositoryAuthBasic:
			vars = append(vars, []corev1.EnvVar{
				{
					Name: "APP_REPOSITORY_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: gitOptions.Auth.SecretName,
							},
							Key: git.UsernameKey,
						},
					},
				},
				{
					Name: "APP_REPOSITORY_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: gitOptions.Auth.SecretName,
							},
							Key: git.PasswordKey,
						},
					},
				},
			}...)
		case git.RepositoryAuthSSHKey:
			vars = append(vars, corev1.EnvVar{
				Name: "APP_REPOSITORY_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: gitOptions.Auth.SecretName,
						},
						Key: git.KeyKey,
					},
				},
			})
			if _, ok := gitOptions.Auth.Credentials[git.PasswordKey]; ok {
				vars = append(vars, corev1.EnvVar{
					Name: "APP_REPOSITORY_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: gitOptions.Auth.SecretName,
							},
							Key: git.PasswordKey,
						},
					},
				})
			}
		}
	}

	return vars
}
