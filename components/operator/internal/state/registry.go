package state

import (
	"context"
	"fmt"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/registry"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	extNamespacedScopeSecretsDetectedFormat = "actual registry configuration in namespace %s comes from %s/%s and it is different from spec.dockerRegistry.secretName. Reflect the %s secret in the secretName field or delete it"
	extRegSecDiffThanSpecFormat             = "actual registry configuration comes from %s/%s and it is different from spec.dockerRegistry.secretName. Reflect the %s secret in the secretName field or delete it"
	extRegSecNotInSpecFormat                = "actual registry configuration comes from %s/%s and it is different from spec.dockerRegistry.secretName. Reflect %s secret in the secretName field"
	internalEnabledAndSecretNameUsedMessage = "spec.dockerRegistry.enableInternal is true and spec.dockerRegistry.secretName is used. Delete the secretName field or set the enableInternal value to false"
	customPVCSMissingInSpec                 = "actual internal registry size is %s and it is different from default value and from spec.dockerRegistry.persistence.size. Configure custom storage size in the spec.dockerRegistry.persistence.size"
	requestedPVCLessThanActual              = "requested storage %s cannot be less than actual storage %s"
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
	extRegSecretClusterWide, err := registry.GetExternalClusterWideRegistrySecret(ctx, r.client, s.instance.GetNamespace())
	if err != nil {
		return err
	}

	extRegSecretNamespacedScope, err := registry.ListExternalNamespacedScopeSecrets(ctx, r.client)
	if err != nil {
		return err
	}

	switch {
	case extRegSecretClusterWide != nil:
		// case: use runtime secret (with labels)
		// doc: https://kyma-project.io/docs/kyma/latest/05-technical-reference/svls-03-switching-registries#cluster-wide-external-registry
		setRuntimeRegistryConfig(extRegSecretClusterWide, s)
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

	addRegistryConfigurationWarnings(extRegSecretClusterWide, extRegSecretNamespacedScope, s)
	return nil
}

func addRegistryConfigurationWarnings(extRegSecretClusterWide *corev1.Secret, extRegSecretsNamespacedScope []corev1.Secret, s *systemState) {
	// runtime secrets (namespaced scope) exist
	for _, secret := range extRegSecretsNamespacedScope {
		s.warningBuilder.With(fmt.Sprintf(extNamespacedScopeSecretsDetectedFormat, secret.Namespace, secret.Namespace, secret.Name, secret.Name))
	}

	// runtime secret (cluster wide) exist and it's other than this under secretName
	if extRegSecretClusterWide != nil && isRegistrySecretName(s.instance.Spec.DockerRegistry) &&
		extRegSecretClusterWide.Name != *s.instance.Spec.DockerRegistry.SecretName {
		s.warningBuilder.With(fmt.Sprintf(extRegSecDiffThanSpecFormat, extRegSecretClusterWide.Namespace, extRegSecretClusterWide.Name, extRegSecretClusterWide.Name))
	}

	// runtime secret exist and secretName field is empty
	if extRegSecretClusterWide != nil && !isRegistrySecretName(s.instance.Spec.DockerRegistry) {
		s.warningBuilder.With(fmt.Sprintf(extRegSecNotInSpecFormat, extRegSecretClusterWide.Namespace, extRegSecretClusterWide.Name, extRegSecretClusterWide.Name))
	}

	// enableInternal is true and secretName is used
	if getEnableInternal(s.instance.Spec.DockerRegistry) && isRegistrySecretName(s.instance.Spec.DockerRegistry) {
		s.warningBuilder.With(internalEnabledAndSecretNameUsedMessage)
	}
}

func setRuntimeRegistryConfig(secret *corev1.Secret, s *systemState) {
	s.instance.Status.DockerRegistry = string(secret.Data["serverAddress"])
}

func setInternalRegistryConfig(ctx context.Context, r *reconciler, s *systemState) error {
	s.instance.Status.DockerRegistry = "internal"
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

	pvcStorage, err := resolvePVCSize(ctx, r, s)
	if err != nil {
		return errors.Wrap(err, "while resolving pvc size")
	}
	if pvcStorage != nil {
		r.log.Debugf("docker registry pvc size: %s", pvcStorage.String())
		s.flagsBuilder.WithRegistryPVSize(pvcStorage.String())
	}

	return nil
}

func resolvePVCSize(ctx context.Context, r *reconciler, s *systemState) (*resource.Quantity, error) {

	actualStorage, err := registry.GetClaimedServerlessDockerRegistryStorageSize(ctx, r.client, s.instance.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "while reading actual PVC size from cluster")
	}

	r.log.Debugf("PVC size : actual %s", actualStorage.String())

	if s.instance.Spec.DockerRegistry.Persistence != nil {
		requestedStorage := s.instance.Spec.DockerRegistry.Persistence.Size
		r.log.Debugf("PVC size : requested %s", requestedStorage.String())

		if actualStorage != nil && requestedStorage.Cmp(*actualStorage) < 0 {
			r.log.Debugf(requestedPVCLessThanActual, requestedStorage.String(), actualStorage.String())
			s.warningBuilder.With(fmt.Sprintf(requestedPVCLessThanActual, requestedStorage.String(), actualStorage.String()))
			return actualStorage, nil
		}
		return &requestedStorage, nil
	} else if actualStorage != nil && !actualStorage.Equal(resource.MustParse("20Gi")) {
		r.log.Debugf("preserving actual pvc size value %s with warning", actualStorage.String())
		s.warningBuilder.With(fmt.Sprintf(customPVCSMissingInSpec, actualStorage.String()))
		return actualStorage, nil
	}
	return nil, nil
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
