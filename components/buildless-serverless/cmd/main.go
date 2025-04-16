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
	"crypto/tls"
	"flag"
	"github.com/kyma-project/serverless/internal/config"
	"github.com/kyma-project/serverless/internal/controller/cache"
	"github.com/vrischmann/envconfig"
	uberzap "go.uber.org/zap"
	uberzapcore "go.uber.org/zap/zapcore"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/controller"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = zap.New().WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(serverlessv1alpha2.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

type serverlessConfig struct {
	MetricsAddress            string `envconfig:"default=:8080"`
	LeaderElectionEnabled     bool   `envconfig:"default=false"`
	LeaderElectionID          string `envconfig:"default=serverless-controller-leader-election-helper"`
	SecretMutatingWebhookPort int    `envconfig:"default=8443"`
	FunctionConfigPath        string `envconfig:"default=hack/function-config.yaml"` // path to development version of function config file
}

func main() {
	cfg, err := loadConfig("APP")
	if err != nil {
		setupLog.Error(err, "unable to load config")
		os.Exit(1)
	}

	functionCfg, err := config.LoadFunctionConfig(cfg.FunctionConfigPath)
	if err != nil {
		setupLog.Error(err, "unable to load configuration file")
		os.Exit(1)
	}

	//TODO: Do we need all these flags?
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	opts := zap.Options{
		// TODO: change this flag for production
		Development: true,
		TimeEncoder: uberzapcore.TimeEncoderOfLayout("Jan 02 15:04:05.000000000"),
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	//TODO: Do we need this part of code (about http2)?
	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: cfg.MetricsAddress,
		},
		LeaderElection:   cfg.LeaderElectionEnabled,
		LeaderElectionID: cfg.LeaderElectionID,
		WebhookServer: webhook.NewServer(webhook.Options{
			Port: cfg.SecretMutatingWebhookPort,
		}),
		HealthProbeBindAddress: probeAddr,
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
		Client: client.Options{
			Cache: &client.CacheOptions{
				DisableFor: []client.Object{
					&serverlessv1alpha2.Function{},
				},
			},
		},
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	logConfig := uberzap.NewDevelopmentConfig()
	logConfig.EncoderConfig.TimeKey = "timestamp"
	logConfig.EncoderConfig.EncodeTime = opts.TimeEncoder
	logConfig.DisableCaller = true

	reconcilerLogger, err := logConfig.Build()
	if err != nil {
		setupLog.Error(err, "unable to setup logger")
		os.Exit(1)
	}

	if err = (&controller.FunctionReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		Log:             reconcilerLogger.Sugar(),
		Config:          functionCfg,
		LastCommitCache: cache.NewRepoLastCommitCache(functionCfg.FunctionReadyRequeueDuration),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Function")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

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
