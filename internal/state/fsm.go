package state

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	defaultResult = ctrl.Result{}
)

type stateFn func(context.Context, *reconciler, *systemState) (stateFn, *ctrl.Result, error)

type cfg struct {
	finalizer string
	chartPath string
}

type systemState struct {
	instance       v1alpha1.Serverless
	dockerRegistry map[string]interface{}
}

func (s *systemState) setState(state v1alpha1.State) {
	s.instance.Status.State = state
}

func (s *systemState) Setup(ctx context.Context, c client.Client) error {
	s.instance.Spec.Default()

	s.dockerRegistry = map[string]interface{}{
		"registryAddress": pointer.String(v1alpha1.DefaultRegistryAddress),
		"serverAddress":   pointer.String(v1alpha1.DefaultServerAddress),
	}
	if s.instance.Spec.DockerRegistry.SecretName != nil {
		var secret corev1.Secret
		key := client.ObjectKey{
			Namespace: s.instance.Namespace,
			Name:      *s.instance.Spec.DockerRegistry.SecretName,
		}
		err := c.Get(ctx, key, &secret)
		if err != nil {
			return err
		}
		for _, k := range []string{"username", "password", "registryAddress", "serverAddress"} {
			if v, ok := secret.Data[k]; ok {
				s.dockerRegistry[k] = string(v)
			}
		}
	}
	return nil
}

type k8s struct {
	client client.Client
	config *rest.Config
}

type reconciler struct {
	fn     stateFn
	log    *zap.SugaredLogger
	cache  *chart.ManifestCache
	result ctrl.Result
	k8s
	cfg
}

func (m *reconciler) stateFnName() string {
	fullName := runtime.FuncForPC(reflect.ValueOf(m.fn).Pointer()).Name()
	splitFullName := strings.Split(fullName, ".")

	if len(splitFullName) < 3 {
		return fullName
	}

	shortName := splitFullName[2]
	return shortName
}

func (m *reconciler) Reconcile(ctx context.Context, v v1alpha1.Serverless) (ctrl.Result, error) {
	state := systemState{instance: v}
	var err error
	var result *ctrl.Result
loop:
	for m.fn != nil && err == nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			break loop

		default:
			m.log.Info(fmt.Sprintf("switching state: %s", m.stateFnName()))
			m.fn, result, err = m.fn(ctx, m, &state)
		}
	}

	if result == nil {
		result = &defaultResult
	}

	m.log.
		With("error", err).
		With("result", result).
		Info("reconciliation done")

	return *result, err
}
