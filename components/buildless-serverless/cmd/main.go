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
	"github.com/go-logr/zapr"
	"os"

	"github.com/kyma-project/serverless/components/buildless-serverless/internal/config"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/cache"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/orphaned-resources"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/logging"
	"github.com/vrischmann/envconfig"
	uberzap "go.uber.org/zap"
	uberzapcore "go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller"
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

	logCfg, err := config.LoadLogConfig(envCfg.LogConfigPath)
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

	log, err := logging.ConfigureLogger(logCfg.LogLevel, logCfg.LogFormat, atomic)
	if err != nil {
		setupLog.Error(err, "unable to configure log")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logWithCtx := log.WithContext()
	go logging.ReconfigureOnConfigChange(ctx, logWithCtx.Named("notifier"), atomic, envCfg.LogConfigPath)

	ctrl.SetLogger(zapr.NewLogger(logWithCtx.Desugar()))

	logWithCtx.Info("Generating Kubernetes client config")
	restConfig := ctrl.GetConfigOrDie()

	logWithCtx.Info("Initializing controller manager")
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: cfg.MetricsAddress,
		},
		LeaderElection:   cfg.LeaderElectionEnabled,
		LeaderElectionID: cfg.LeaderElectionID,
		WebhookServer: webhook.NewServer(webhook.Options{
			Port: cfg.SecretMutatingWebhookPort,
		}),
		HealthProbeBindAddress: cfg.HealthzAddress,
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

	if err = (&controller.FunctionReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		Log:             logWithCtx,
		Config:          cfg,
		LastCommitCache: cache.NewRepoLastCommitCache(cfg.FunctionReadyRequeueDuration),
		EventRecorder:   mgr.GetEventRecorderFor(serverlessv1alpha2.FunctionControllerValue),
	}).SetupWithManager(mgr); err != nil {
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

func loadConfig(prefix string) (serverlessConfig, error) {
	cfg := serverlessConfig{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
