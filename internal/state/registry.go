package state

import (
	"context"
	"fmt"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/kyma-project/serverless-manager/internal/registry"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func sFnRegistryConfiguration(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	// setup status.dockerRegistry and set possible warnings
	err := configureRegistry(ctx, r, s)
	if err != nil {
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigurationErr,
			err,
		)
		return nextState(sFnUpdateStatusWithError(err))
	}

	// update status if needed
	if s.snapshot.DockerRegistry != s.instance.Status.DockerRegistry {
		return nextState(sFnUpdateStatusAndRequeue)
	}

	return nextState(sFnOptionalDependencies)
}

func configureRegistry(ctx context.Context, r *reconciler, s *systemState) error {
	secret, err := registry.GetServerlessExternalRegistrySecret(ctx, r.client, s.instance.GetNamespace())
	if err != nil {
		return err
	}

	switch {
	case secret != nil:
		// case: use runtime secret (with labels)
		// doc: https://kyma-project.io/docs/kyma/latest/05-technical-reference/svls-03-switching-registries#cluster-wide-external-registry
		setRuntimeRegistryConfig(secret, s)
	case s.instance.Spec.DockerRegistry.SecretName != nil:
		// case: use secret from secretName field
		err := setExternalRegistryConfig(ctx, r, s)
		if err != nil {
			return err
		}
	case *s.instance.Spec.DockerRegistry.EnableInternal:
		// case: use internal registry
		err := setInternalRegistryConfig(ctx, r, s)
		if err != nil {
			return err
		}
	default:
		// case: use k3d registry
		setK3dRegistryConfig(s)
	}

	addRegistryConfigurationWarnings(secret, s)
	return nil
}

func addRegistryConfigurationWarnings(secret *corev1.Secret, s *systemState) {
	// runtime secret exist and it's other than this under secretName
	if secret != nil &&
		s.instance.Spec.DockerRegistry.SecretName != nil &&
		secret.Name != *s.instance.Spec.DockerRegistry.SecretName {
		s.addWarning(fmt.Sprintf("actual registry configuration comes from %s/%s and it's different from spec.dockerRegistry.secretName. Reflect the %s secret in the secretName field or delete it", secret.Name, secret.Namespace, secret.Name))
	}

	// runtime secret exist and secretName field is empty
	if secret != nil && s.instance.Spec.DockerRegistry.SecretName == nil {
		s.addWarning(fmt.Sprintf("actual registry configuration comes from %s/%s and it's different from spec.dockerRegistry.secretName. Reflect %s secret in the secretName field", secret.Name, secret.Namespace, secret.Name))
	}

	// enableInternal is true and secretName is used
	if *s.instance.Spec.DockerRegistry.EnableInternal && s.instance.Spec.DockerRegistry.SecretName != nil {
		s.addWarning("spec.dockerRegistry.enableInternal is true and spec.dockerRegistry.secretName is used. Delete the secretName field or set the enableInternal value to false")
	}
}

func setRuntimeRegistryConfig(secret *corev1.Secret, s *systemState) {
	s.instance.Status.DockerRegistry = string(secret.Data["serverAddress"])
}

func setInternalRegistryConfig(ctx context.Context, r *reconciler, s *systemState) error {
	s.instance.Status.DockerRegistry = "internal"
	s.chartConfig.Release.Flags = chart.AppendInternalRegistryFlags(
		s.chartConfig.Release.Flags,
		*s.instance.Spec.DockerRegistry.EnableInternal,
	)

	resolver := registry.NewNodePortResolver(registry.RandomNodePort)
	nodePort, err := resolver.ResolveDockerRegistryNodePortFn(ctx, r.client, s.instance.Namespace)
	if err != nil {
		return errors.Wrap(err, "while resolving registry node port")
	}
	r.log.Debugf("docker registry node port: %d", nodePort)
	s.chartConfig.Release.Flags = chart.AppendNodePortFlag(s.chartConfig.Release.Flags, int64(nodePort))
	return nil
}

func setExternalRegistryConfig(ctx context.Context, r *reconciler, s *systemState) error {
	secret, err := getRegistrySecret(ctx, r, s)
	if err != nil {
		return err
	}

	s.instance.Status.DockerRegistry = string(secret.Data["serverAddress"])
	s.chartConfig.Release.Flags = chart.AppendExternalRegistryFlags(
		s.chartConfig.Release.Flags,
		*s.instance.Spec.DockerRegistry.EnableInternal,
		string(secret.Data["username"]),
		string(secret.Data["password"]),
		string(secret.Data["registryAddress"]),
		s.instance.Status.DockerRegistry,
	)
	return nil
}

func setK3dRegistryConfig(s *systemState) {
	s.instance.Status.DockerRegistry = v1alpha1.DefaultServerAddress
	s.chartConfig.Release.Flags = chart.AppendK3dRegistryFlags(
		s.chartConfig.Release.Flags,
		*s.instance.Spec.DockerRegistry.EnableInternal,
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
