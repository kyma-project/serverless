package chart

import (
	"context"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// build flags for the installation. The Serverless resource should be validated and ready to use
func BuildFlags(ctx context.Context, client client.Client, serverless *v1alpha1.Serverless) (map[string]interface{}, error) {
	dockerRegistryFlags, err := dockerRegistryFlags(ctx, client, serverless)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"dockerRegistry": dockerRegistryFlags,
	}, nil
}

func AppendContainersFlags(flags map[string]interface{}, publisherURL, traceCollectorURL string) map[string]interface{} {
	flags["containers"] = map[string]interface{}{
		"manager": map[string]interface{}{
			"envs": map[string]interface{}{
				"functionTraceCollectorEndpoint": map[string]interface{}{
					"value": traceCollectorURL,
				},
				"functionPublisherProxyAddress": map[string]interface{}{
					"value": publisherURL,
				},
			},
		},
	}

	return flags
}

func dockerRegistryFlags(ctx context.Context, c client.Client, serverless *v1alpha1.Serverless) (map[string]interface{}, error) {
	flags := map[string]interface{}{
		"enableInternal":  *serverless.Spec.DockerRegistry.EnableInternal,
		"registryAddress": v1alpha1.DefaultRegistryAddress,
		"serverAddress":   v1alpha1.DefaultServerAddress,
	}

	if serverless.Spec.DockerRegistry.SecretName != nil {
		var secret corev1.Secret
		key := client.ObjectKey{
			Namespace: serverless.Namespace,
			Name:      *serverless.Spec.DockerRegistry.SecretName,
		}
		err := c.Get(ctx, key, &secret)
		if err != nil {
			return nil, err
		}
		for _, k := range []string{"username", "password", "registryAddress", "serverAddress"} {
			if v, ok := secret.Data[k]; ok {
				flags[k] = string(v)
			}
		}
	}

	return flags, nil
}
