package state

import (
	"context"
	"fmt"
	"github.com/kyma-project/serverless-manager/internal/registry"
	"github.com/pkg/errors"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func sFnRegistryConfiguration(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {

	secret, err := registry.GetServerlessExternalRegistrySecret(ctx, r.client, s.instance.GetNamespace())
	if err != nil {
		s.setState(v1alpha1.StateError)
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigurationErr,
			err,
		)
		return nextState(sFnUpdateStatusWithError(err))
	}

	switch {
	case secret != nil:
		setExternalRegistrySecretNameInDockerRegistryStatus(secret, s)
	case s.instance.Spec.DockerRegistry.SecretName != nil:
		err := setExternalRegistry(ctx, r, s)
		if err != nil {
			s.setState(v1alpha1.StateError)
			s.instance.UpdateConditionFalse(
				v1alpha1.ConditionTypeConfigured,
				v1alpha1.ConditionReasonConfigurationErr,
				err,
			)
			return nextState(sFnUpdateStatusWithError(err))
		}
	case *s.instance.Spec.DockerRegistry.EnableInternal:
		err := setInternalRegistry(ctx, r, s)
		if err != nil {
			s.setState(v1alpha1.StateError)
			s.instance.UpdateConditionFalse(
				v1alpha1.ConditionTypeConfigured,
				v1alpha1.ConditionReasonConfigurationErr,
				err,
			)
			return nextState(sFnUpdateStatusWithError(err))
		}
	default:
		setK3dRegistry(s)
	}

	setRegistryConfigurationWarnings(secret, s)

	if s.snapshot.DockerRegistry != s.instance.Status.DockerRegistry {
		return nextState(sFnUpdateStatusAndRequeue)
	}

	return nextState(sFnOptionalDependencies)
}

func setRegistryConfigurationWarnings(secret *corev1.Secret, s *systemState) {
	if secret != nil && (s.instance.Spec.DockerRegistry.SecretName == nil || secret.Name != *s.instance.Spec.DockerRegistry.SecretName) {
		s.addWarning(fmt.Sprintf("used registry coming from secret %s/%s, please fill the field spec.dockerRegistry.secretName to match configured secret ", secret.Name, secret.Namespace))
	}
	if secret != nil && *s.instance.Spec.DockerRegistry.EnableInternal == true {
		s.addWarning(fmt.Sprintf("used registry coming from secret %s/%s, spec.dockerRegistry.enableInternal is enabled but not used - edit spec or delete secret to match desired state ", secret.Name, secret.Namespace))
	}
	if *s.instance.Spec.DockerRegistry.EnableInternal == true && s.instance.Spec.DockerRegistry.SecretName != nil {
		s.addWarning("both spec.dockerRegistry.enableInternal is enabled & spec.dockerRegistry.secretName exists - delete secretName field or set enableInternal value to false ")
	}
}

func setExternalRegistrySecretNameInDockerRegistryStatus(secret *corev1.Secret, s *systemState) {
	// doc: https://kyma-project.io/docs/kyma/latest/05-technical-reference/svls-03-switching-registries#cluster-wide-external-registry
	if secret == nil {
		return
	}
	if address, ok := secret.Data["serverAddress"]; ok {
		s.instance.Status.DockerRegistry = string(address)
	}
	return
}

func setInternalRegistry(ctx context.Context, r *reconciler, s *systemState) error {
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

func setExternalRegistry(ctx context.Context, r *reconciler, s *systemState) error {
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

func setK3dRegistry(s *systemState) {
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
