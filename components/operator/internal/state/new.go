package state

import (
	"context"
	"os"

	//TODO: change to "github.com/kyma-project/serverless-manager/components/operator/api/v1alpha1" after new SO structure will be on github
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	//TODO: change to "github.com/kyma-project/serverless-manager/components/operator/internal/chart" after new SO structure will be on github
	"github.com/kyma-project/serverless-manager/internal/chart"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StateReconciler interface {
	Reconcile(ctx context.Context, v v1alpha1.Serverless) (ctrl.Result, error)
}

func NewMachine(client client.Client, config *rest.Config, recorder record.EventRecorder, log *zap.SugaredLogger, cache chart.ManifestCache, chartPath, namespace string) StateReconciler {
	return &reconciler{
		fn:    sFnServedFilter,
		cache: cache,
		log:   log,
		cfg: cfg{
			finalizer:     v1alpha1.Finalizer,
			chartPath:     chartPath,
			namespace:     namespace,
			managerPodUID: os.Getenv("SERVERLESS_MANAGER_UID"),
		},
		k8s: k8s{
			client:        client,
			config:        config,
			EventRecorder: recorder,
		},
	}
}
