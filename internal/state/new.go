package state

import (
	"context"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"go.uber.org/zap"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StateReconciler interface {
	Reconcile(ctx context.Context, v v1alpha1.Serverless) (ctrl.Result, error)
}

func NewMachine(client client.Client, log *zap.SugaredLogger, cache *chart.RendererCache, chartPath string) StateReconciler {
	return &reconciler{
		fn:    sFnInitialize,
		cache: cache,
		log:   log,
		cfg: cfg{
			finalizer: v1alpha1.Finalizer,
			chartPath: chartPath,
		},
		k8s: k8s{
			client: client,
		},
	}
}
