/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/kyma-project/serverless/components/operator/internal/config"
	"github.com/kyma-project/serverless/components/operator/internal/gitrepository"
	"github.com/pkg/errors"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	uberzap "go.uber.org/zap"
	uberzapcore "go.uber.org/zap/zapcore"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	corev1 "k8s.io/api/core/v1"
	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	ctrlwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	operatorv1alpha1 "github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/controllers"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	//+kubebuilder:scaffold:imports
)

var (
	scheme         = runtime.NewScheme()
	setupLog       = ctrl.Log.WithName("setup")
	syncPeriod     = time.Minute * 30
	cleanupTimeout = time.Second * 10
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))

	utilruntime.Must(apiextensionsscheme.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")

	opts := zap.Options{
		Development: true,
		TimeEncoder: uberzapcore.TimeEncoderOfLayout("Jan 02 15:04:05.000000000"),
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	cfg, err := config.GetConfig("")
	if err != nil {
		setupLog.Error(err, "while getting config")
		os.Exit(1)
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	setupLog.Info("cleaning orphan deprecated resources")
	err = cleanupOrphanDeprecatedResources(ctx)
	if err != nil {
		setupLog.Error(err, "while removing orphan resources")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: ctrlmetrics.Options{
			BindAddress: metricsAddr,
		},
		WebhookServer: ctrlwebhook.NewServer(ctrlwebhook.Options{
			Port: 9443,
		}),
		HealthProbeBindAddress: probeAddr,
		Cache: ctrlcache.Options{
			SyncPeriod: &syncPeriod,
		},
		Client: ctrlclient.Options{
			Cache: &ctrlclient.CacheOptions{
				DisableFor: []ctrlclient.Object{
					&corev1.Secret{},
					&corev1.ConfigMap{},
				},
			},
		},
		// TODO: use our own logger - now eventing use logger with different message format
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	config := uberzap.NewDevelopmentConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = opts.TimeEncoder
	config.DisableCaller = true

	reconcilerLogger, err := config.Build()
	if err != nil {
		setupLog.Error(err, "unable to setup logger")
		os.Exit(1)
	}

	reconciler := controllers.NewServerlessReconciler(
		mgr.GetClient(), mgr.GetConfig(),
		mgr.GetEventRecorderFor("serverless-operator"),
		reconcilerLogger.Sugar(),
		cfg.ChartPath)

	if err = reconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Serverless")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func cleanupOrphanDeprecatedResources(ctx context.Context) error {
	// We are going to talk to the API server _before_ we start the manager.
	// Since the default manager client reads from cache, we will get an error.
	// So, we create a "serverClient" that would read from the API directly.
	// We only use it here, this only runs at start up, so it shouldn't be to much for the API
	serverClient, err := ctrlclient.New(ctrl.GetConfigOrDie(), ctrlclient.Options{
		Scheme: scheme,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create a server client")
	}

	return gitrepository.Cleanup(ctx, serverClient)
}
