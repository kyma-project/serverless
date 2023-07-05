package state

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func sFnRegistryConfiguration(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	switch {
	case *s.instance.Spec.DockerRegistry.EnableInternal:
		setInternalRegistry(s)
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
	default:
		setK3dRegistry(s)
	}

	if condition, ok := s.instance.GetCondition(v1alpha1.ConditionTypeConfigured); ok {
		if condition.Status == metav1.ConditionFalse {
			s.setState(v1alpha1.StateReady)
			s.instance.UpdateConditionTrue(
				v1alpha1.ConditionTypeConfigured,
				v1alpha1.ConditionReasonConfigured,
				"Configured",
			)
			return nextState(sFnUpdateStatusAndRequeue)
		}
	}

	if s.snapshot.DockerRegistry != s.instance.Status.DockerRegistry {
		return nextState(sFnUpdateStatusAndRequeue)
	}

	return nextState(sFnOptionalDependencies)
}

func setInternalRegistry(s *systemState) {
	s.instance.Status.DockerRegistry = "internal"
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
