package state

import (
	"context"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/registry"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	extNamespacedScopeSecretsDetectedFormat = "actual registry configuration in namespace %s comes from %s/%s and it is different from spec.dockerRegistry.secretName. Reflect the %s secret in the secretName field or delete it"
	extRegSecDiffThanSpecFormat             = "actual registry configuration comes from %s/%s and it is different from spec.dockerRegistry.secretName. Reflect the %s secret in the secretName field or delete it"
	extRegSecNotInSpecFormat                = "actual registry configuration comes from %s/%s and it is different from spec.dockerRegistry.secretName. Reflect %s secret in the secretName field"
	internalEnabledAndSecretNameUsedMessage = "spec.dockerRegistry.enableInternal is true and spec.dockerRegistry.secretName is used. Delete the secretName field or set the enableInternal value to false"
)

func sFnRegistryConfiguration(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	s.setState(v1alpha1.StateProcessing)
	// setup status.dockerRegistry and set possible warnings
	err := configureRegistry(ctx, r, s)
	if err != nil {
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigurationErr,
			err,
		)
		return stopWithEventualError(err)
	}

	return nextState(sFnOptionalDependencies)
}

func configureRegistry(ctx context.Context, r *reconciler, s *systemState) error {

	switch {
	case isRegistrySecretName(s.instance.Spec.DockerRegistry):
		// case: use secret from secretName field
		err := setExternalRegistryConfig(ctx, r, s)
		if err != nil {
			return err
		}
	case getEnableInternal(s.instance.Spec.DockerRegistry):
		// case: use internal registry
		err := setInternalRegistryConfig(ctx, r, s)
		if err != nil {
			return err
		}
	default:
		// case: use k3d registry
		setK3dRegistryConfig(s)
	}
	addRegistryConfigurationWarnings(s)
	return nil
}

func addRegistryConfigurationWarnings(s *systemState) {

	// enableInternal is true and secretName is used
	if getEnableInternal(s.instance.Spec.DockerRegistry) && isRegistrySecretName(s.instance.Spec.DockerRegistry) {
		s.warningBuilder.With(internalEnabledAndSecretNameUsedMessage)
	}
}

func setInternalRegistryConfig(ctx context.Context, r *reconciler, s *systemState) error {
	// TODO: this is a temporary solution, delete it after removing legacy serverless
	if isLegacyEnabled(s.instance.Annotations) {
		s.instance.Status.DockerRegistry = "internal"
	} else {
		s.instance.Status.DockerRegistry = ""
	}
	s.flagsBuilder.WithRegistryEnableInternal(
		*s.instance.Spec.DockerRegistry.EnableInternal,
	)

	existingIntRegSecret, err := registry.GetServerlessInternalRegistrySecret(ctx, r.client, s.instance.Namespace)
	if err != nil {
		return errors.Wrap(err, "while fetching existing serverless internal docker registry secret")
	}
	if existingIntRegSecret != nil {
		r.log.Debugf("reusing existing credentials for internal docker registry to avoiding docker registry  rollout")
		registryHttpSecretEnvValue, getErr := registry.GetRegistryHTTPSecretEnvValue(ctx, r.client, s.instance.Namespace)
		if getErr != nil {
			return errors.Wrap(getErr, "while reading env value registryHttpSecret from serverless internal docker registry deployment")
		}
		s.flagsBuilder.
			WithRegistryCredentials(
				string(existingIntRegSecret.Data["username"]),
				string(existingIntRegSecret.Data["password"]),
			).
			WithRegistryHttpSecret(
				registryHttpSecretEnvValue,
			)
	}

	resolver := registry.NewNodePortResolver(registry.RandomNodePort)
	nodePort, err := resolver.ResolveDockerRegistryNodePortFn(ctx, r.client, s.instance.Namespace)
	if err != nil {
		return errors.Wrap(err, "while resolving registry node port")
	}
	r.log.Debugf("docker registry node port: %d", nodePort)
	s.flagsBuilder.WithNodePort(int64(nodePort))
	return nil
}

func setExternalRegistryConfig(ctx context.Context, r *reconciler, s *systemState) error {
	secret, err := getRegistrySecret(ctx, r, s)
	if err != nil {
		return err
	}

	s.instance.Status.DockerRegistry = string(secret.Data["serverAddress"])
	s.flagsBuilder.
		WithRegistryEnableInternal(
			getEnableInternal(s.instance.Spec.DockerRegistry),
		).
		WithRegistryCredentials(
			string(secret.Data["username"]),
			string(secret.Data["password"]),
		).
		WithRegistryAddresses(
			string(secret.Data["registryAddress"]),
			s.instance.Status.DockerRegistry,
		)

	return nil
}

func setK3dRegistryConfig(s *systemState) {
	s.instance.Status.DockerRegistry = v1alpha1.DefaultServerAddress
	s.flagsBuilder.WithRegistryEnableInternal(
		getEnableInternal(s.instance.Spec.DockerRegistry),
	).WithRegistryAddresses(
		v1alpha1.DefaultRegistryAddress,
		s.instance.Status.DockerRegistry,
	)
}

func getRegistrySecret(ctx context.Context, r *reconciler, s *systemState) (*corev1.Secret, error) {
	var secret corev1.Secret
	key := client.ObjectKey{
		Namespace: s.instance.Namespace,
		Name:      *s.instance.Spec.DockerRegistry.SecretName,
	}
	err := r.client.Get(ctx, key, &secret)
	return &secret, err
}

func isRegistrySecretName(registry *v1alpha1.DockerRegistry) bool {
	return registry != nil && registry.SecretName != nil
}

func getEnableInternal(registry *v1alpha1.DockerRegistry) bool {
	if registry != nil && registry.EnableInternal != nil {
		return *registry.EnableInternal
	}
	return v1alpha1.DefaultEnableInternal
}
