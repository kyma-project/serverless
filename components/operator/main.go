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
	"crypto/fips140"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/serverless/components/operator/internal/logging"
	"github.com/vrischmann/envconfig"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"github.com/go-logr/zapr"
	logconfig "github.com/kyma-project/manager-toolkit/logging/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	ctrlwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	operatorv1alpha1 "github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/controllers"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	//+kubebuilder:scaffold:imports
)

var (
	scheme     = runtime.NewScheme()
	setupLog   logr.Logger
	syncPeriod = time.Minute * 30
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))

	utilruntime.Must(apiextensionsscheme.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

type operatorConfig struct {
	ChartPath     string `envconfig:"default=/module-chart"`
	LogConfigPath string `envconfig:"default=hack/log-config.yaml"`
}

func main() {
	if !isFIPS140Only() {
		fmt.Printf("FIPS 140 exclusive mode is not enabled. Check GODEBUG flags. FIPS not enforced\n")
		panic("FIPS 140 exclusive mode is not enabled. Check GODEBUG flags.")
	}

	var metricsAddr string
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.Parse()

	// Load operator configuration from environment variables
	opCfg, err := loadConfig("")
	if err != nil {
		fmt.Printf("unable to load config: %v\n", err)
		os.Exit(1)
	}

	// Load log configuration from file
	logCfg, err := logconfig.LoadConfig(opCfg.LogConfigPath)
	if err != nil {
		fmt.Printf("unable to load log configuration file: %v\n", err)
		os.Exit(1)
	}

	logLevel := logCfg.LogLevel
	logFormat := logCfg.LogFormat

	atomicLevel := zap.NewAtomicLevel()
	parsedLevel, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		fmt.Printf("unable to parse logger level: %v\n", err)
		os.Exit(1)
	}
	atomicLevel.SetLevel(parsedLevel)

	// Configure logger using manager-toolkit
	log, err := logging.ConfigureLogger(logLevel, logFormat, atomicLevel)
	if err != nil {
		fmt.Printf("unable to configure logger: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logWithCtx := log.WithContext()

	// Set initial format for change detection (pod will auto-restart on format changes)
	logconfig.SetInitialFormat(logFormat)

	// Start log config watcher with restart callback
	go logging.ReconfigureOnConfigChangeWithRestart(ctx, logWithCtx.Named("notifier"), atomicLevel, opCfg.LogConfigPath, func() {
		// Trigger graceful restart by exiting after a short delay
		go func() {
			time.Sleep(2 * time.Second)
			logWithCtx.Info("Exiting for pod restart due to log format change")
			os.Exit(0)
		}()
	})

	ctrl.SetLogger(zapr.NewLogger(logWithCtx.Desugar()))
	setupLog = ctrl.Log.WithName("setup")

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
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	reconciler := controllers.NewServerlessReconciler(
		mgr.GetClient(), mgr.GetConfig(),
		mgr.GetEventRecorderFor("serverless-operator"),
		logWithCtx.Desugar().Sugar(),
		opCfg.ChartPath)

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

func isFIPS140Only() bool {
	return fips140.Enabled() && os.Getenv("GODEBUG") == "fips140=only,tlsmlkem=0"
}

func loadConfig(prefix string) (operatorConfig, error) {
	cfg := operatorConfig{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
