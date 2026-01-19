/*
Copyright 2024.

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
	"io"
	"log"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/go-logr/zapr"
	logconfig "github.com/kyma-project/manager-toolkit/logging/config"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/config"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/git"
	serverlessmetrics "github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/metrics"
	orphaned_resources "github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/orphaned-resources"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/endpoint"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/logging"
	"github.com/vrischmann/envconfig"
	uberzap "go.uber.org/zap"
	uberzapcore "go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrlzap.New().WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(serverlessv1alpha2.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

type serverlessConfig struct {
	FunctionConfigPath string `envconfig:"default=hack/function-config.yaml"` // path to development version of function config file
	LogConfigPath      string `envconfig:"default=hack/log-config.yaml"`      // path to development version of log config file
}

func main() {

	envCfg, err := loadConfig("APP")
	if err != nil {
		setupLog.Error(err, "unable to load config")
		os.Exit(1)
	}

	cfg, err := config.LoadFunctionConfig(envCfg.FunctionConfigPath)
	if err != nil {
		setupLog.Error(err, "unable to load function configuration file")
		os.Exit(1)
	}

	logCfg, err := logconfig.LoadConfig(envCfg.LogConfigPath)
	if err != nil {
		setupLog.Error(err, "unable to load log configuration file")
		os.Exit(1)
	}

	atomic := uberzap.NewAtomicLevel()
	parsedLevel, err := uberzapcore.ParseLevel(logCfg.LogLevel)
	if err != nil {
		setupLog.Error(err, "unable to parse logger level")
		os.Exit(1)
	}
	atomic.SetLevel(parsedLevel)

	logger, err := logging.ConfigureLogger(logCfg.LogLevel, logCfg.LogFormat, atomic)
	if err != nil {
		setupLog.Error(err, "unable to configure log")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logWithCtx := logger.WithContext()

	// Set initial format for change detection (pod will auto-restart on format changes)
	logconfig.SetInitialFormat(logCfg.LogFormat)

	go logging.ReconfigureOnConfigChange(ctx, logWithCtx.Named("notifier"), atomic, envCfg.LogConfigPath)

	ctrl.SetLogger(zapr.NewLogger(logWithCtx.Desugar()))

	logWithCtx.Info("Generating Kubernetes client config")
	restConfig := ctrl.GetConfigOrDie()

	logWithCtx.Info("Initializing controller manager")
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: cfg.MetricsPort,
		},
		LeaderElection:   cfg.LeaderElectionEnabled,
		LeaderElectionID: cfg.LeaderElectionID,
		WebhookServer: webhook.NewServer(webhook.Options{
			Port: cfg.SecretMutatingWebhookPort,
		}),
		HealthProbeBindAddress: cfg.Healthz.Port,
		Client: client.Options{
			Cache: &client.CacheOptions{
				DisableFor: []client.Object{
					&serverlessv1alpha2.Function{},
					&corev1.Secret{},
					&corev1.ConfigMap{},
				},
			},
		},
		// TODO: add cache config if needed
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	serverlessmetrics.Register()

	healthHandler, healthEventsCh, healthResponseCh := controller.NewHealthChecker(cfg.Healthz.LivenessTimeout, logWithCtx.Named("healthz"))
	if err := mgr.AddHealthzCheck("healthz", healthHandler.Checker); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	fnCtrl, err := (&controller.FunctionReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		Log:           logWithCtx,
		Config:        cfg,
		EventRecorder: mgr.GetEventRecorderFor(serverlessv1alpha2.FunctionControllerValue),
		GitChecker:    git.NewAsyncLatestCommitChecker(ctx, logWithCtx),
		HealthCh:      healthResponseCh,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Function")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	//TODO: It is a temporary solution to delete orphaned jobs. It should be removed after migration from old serverless
	go func() {
		err := orphaned_resources.DeleteOrphanedResources(ctx, mgr)
		if err != nil {
			setupLog.Error(err, "unable to delete jobs")
		}
	}()

	err = fnCtrl.Watch(source.Channel(healthEventsCh, &handler.EnqueueRequestForObject{}))
	if err != nil {
		setupLog.Error(err, "unable to watch health events channel")
		os.Exit(1)
	}

	// disable default log to prevent http server from logging returned status codes
	log.SetOutput(io.Discard)

	internalServer := endpoint.NewInternalServer(ctx, logWithCtx, mgr.GetClient(), cfg)
	go func() {
		err := internalServer.ListenAndServe(cfg.InternalEndpointPort)
		if err != nil {
			logWithCtx.Error(err, "internal HTTP server error")
		}
	}()

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func loadConfig(prefix string) (serverlessConfig, error) {
	cfg := serverlessConfig{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
