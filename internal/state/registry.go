package state

import (
	"context"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	corev1 "k8s.io/api/core/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// k3d
// enableInternal: false
// secretName: nil

// internal
// enableInternal: true
// secretName: nil

// external
// enableInternal: false
// secretName: <secret-name>

func sFnRegistryConfiguration() stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *controllerruntime.Result, error) {
		switch {
		case *s.instance.Spec.DockerRegistry.EnableInternal:
			setInternalRegistry(s)
		case s.instance.Spec.DockerRegistry.SecretName != nil:
			err := setExternalRegistry(ctx, r, s)
			if err != nil {
				return nextState(
					sFnUpdateErrorState(
						v1alpha1.ConditionTypeConfigured,
						v1alpha1.ConditionReasonConfigurationErr,
						err,
					),
				)
			}
		default:
			setK3dRegistry(s)
		}

		if s.snapshot.Registry != s.instance.Status.Registry {
			return nextState(
				sFnUpdateStatusWithRequeue,
			)
		}

		return nextState(
			sFnOptionalDependencies(),
		)
	}
}

func setInternalRegistry(s *systemState) {
	s.instance.Status.Registry = "internal"
	s.chartConfig.Release.Flags = chart.AppendInternalRegistryFlags(
		s.chartConfig.Release.Flags,
		*s.instance.Spec.DockerRegistry.EnableInternal,
	)
}

func setExternalRegistry(ctx context.Context, r *reconciler, s *systemState) error {
	secret, err := getRegistrySecret(ctx, r, s)
	if err != nil {
		return err
	}

	s.instance.Status.Registry = string(secret.Data["serverAddress"])
	s.chartConfig.Release.Flags = chart.AppendExternalRegistryFlags(
		s.chartConfig.Release.Flags,
		*s.instance.Spec.DockerRegistry.EnableInternal,
		string(secret.Data["username"]),
		string(secret.Data["password"]),
		string(secret.Data["registryAddress"]),
		s.instance.Status.Registry,
	)
	return nil
}

func setK3dRegistry(s *systemState) {
	s.instance.Status.Registry = v1alpha1.DefaultServerAddress
	s.chartConfig.Release.Flags = chart.ApendK3dRegistryFlags(
		s.chartConfig.Release.Flags,
		*s.instance.Spec.DockerRegistry.EnableInternal,
		v1alpha1.DefaultRegistryAddress,
		s.instance.Status.Registry,
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
