package state

import (
	"context"
	"os"

	"github.com/kyma-project/manager-toolkit/installation/chart"
	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StateReconciler interface {
	Reconcile(ctx context.Context, v v1alpha1.Serverless) (ctrl.Result, error)
}

func NewMachine(client client.Client, config *rest.Config, recorder record.EventRecorder, log *zap.SugaredLogger, cache chart.ManifestCache, chartPath string) StateReconciler {
	return &reconciler{
		fn:    sFnServedFilter,
		cache: cache,
		log:   log,
		cfg: cfg{
			finalizer:     v1alpha1.Finalizer,
			chartPath:     chartPath,
			managerPodUID: os.Getenv("SERVERLESS_MANAGER_UID"),
		},
		k8s: k8s{
			client:        client,
			config:        config,
			EventRecorder: recorder,
		},
	}
}
