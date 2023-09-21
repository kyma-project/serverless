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

const (
	extRegSecDiffThanSpecFormat             = "actual registry configuration comes from %s/%s and it is different from spec.dockerRegistry.secretName. Reflect the %s secret in the secretName field or delete it"
	extRegSecNotInSpecFormat                = "actual registry configuration comes from %s/%s and it is different from spec.dockerRegistry.secretName. Reflect %s secret in the secretName field"
	internalEnabledAndSecretNameUsedMessage = "spec.dockerRegistry.enableInternal is true and spec.dockerRegistry.secretName is used. Delete the secretName field or set the enableInternal value to false"
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
		return stopWithPossibleError(err)
	}

	return nextState(sFnOptionalDependencies)
}

func configureRegistry(ctx context.Context, r *reconciler, s *systemState) error {
	extRegSecert, err := registry.GetServerlessExternalRegistrySecret(ctx, r.client, s.instance.GetNamespace())
	if err != nil {
		return err
	}

	switch {
	case extRegSecert != nil:
		// case: use runtime secret (with labels)
		// doc: https://kyma-project.io/docs/kyma/latest/05-technical-reference/svls-03-switching-registries#cluster-wide-external-registry
		setRuntimeRegistryConfig(extRegSecert, s)
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

	addRegistryConfigurationWarnings(extRegSecert, s)
	return nil
}

func addRegistryConfigurationWarnings(extRegSecert *corev1.Secret, s *systemState) {
	// runtime secret exist and it's other than this under secretName
	if extRegSecert != nil &&
		s.instance.Spec.DockerRegistry.SecretName != nil &&
		extRegSecert.Name != *s.instance.Spec.DockerRegistry.SecretName {
		s.warningBuilder.With(fmt.Sprintf(extRegSecDiffThanSpecFormat, extRegSecert.Namespace, extRegSecert.Name, extRegSecert.Name))
	}

	// runtime secret exist and secretName field is empty
	if extRegSecert != nil && s.instance.Spec.DockerRegistry.SecretName == nil {
		s.warningBuilder.With(fmt.Sprintf(extRegSecNotInSpecFormat, extRegSecert.Namespace, extRegSecert.Name, extRegSecert.Name))
	}

	// enableInternal is true and secretName is used
	if *s.instance.Spec.DockerRegistry.EnableInternal && s.instance.Spec.DockerRegistry.SecretName != nil {
		s.warningBuilder.With(internalEnabledAndSecretNameUsedMessage)
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
		s.chartConfig.Release.Flags = chart.AppendExistingInternalRegistryCredentialsFlags(s.chartConfig.Release.Flags, string(existingIntRegSecret.Data["username"]), string(existingIntRegSecret.Data["password"]), registryHttpSecretEnvValue)
	}

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
