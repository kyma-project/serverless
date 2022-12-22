package state

import (
	"context"

	"github.com/kyma-project/module-manager/pkg/types"
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	chartNs         = "kyma-system"
	createNamespace = false // TODO: think about it
)

type StateReconciler interface {
	Reconcile(ctx context.Context, v v1alpha1.Serverless) (ctrl.Result, error)
}

func NewMachine(client client.Client, config *rest.Config, log *zap.SugaredLogger, cache types.CacheManager, chartPath string) StateReconciler {
	return &reconciler{
		fn:           sFnInitialize,
		cacheManager: cache,
		log:          log,
		cfg: cfg{
			finalizer:       v1alpha1.Finalizer,
			chartPath:       chartPath,
			chartNs:         chartNs,
			createNamespace: createNamespace,
		},
		k8s: k8s{
			client: client,
			config: config,
		},
	}
}
