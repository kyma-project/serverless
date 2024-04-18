package state

import (
	"context"
	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/registry"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
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

	return nextState(sFnControllerConfiguration)
}

func configureRegistry(ctx context.Context, r *reconciler, s *systemState) error {
	err := setInternalRegistryConfig(ctx, r, s)
	if err != nil {
		return err
	}

	return nil
}

func setInternalRegistryConfig(ctx context.Context, r *reconciler, s *systemState) error {
	existingIntRegSecret, err := registry.GetDockerRegistryInternalRegistrySecret(ctx, r.client, s.instance.Namespace)
	if err != nil {
		return errors.Wrap(err, "while fetching existing internal docker registry secret")
	}
	if existingIntRegSecret != nil {
		r.log.Debugf("reusing existing credentials for internal docker registry to avoiding docker registry  rollout")
		registryHttpSecretEnvValue, getErr := registry.GetRegistryHTTPSecretEnvValue(ctx, r.client, s.instance.Namespace)
		if getErr != nil {
			return errors.Wrap(getErr, "while reading env value registryHttpSecret from internal docker registry deployment")
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
