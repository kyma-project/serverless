package main

import (
	"context"
	"crypto/fips140"
	"os"

	"github.com/go-logr/zapr"
	logconfig "github.com/kyma-project/manager-toolkit/logging/config"
	"github.com/kyma-project/serverless/components/serverless/internal/logging"
	"github.com/kyma-project/serverless/components/serverless/internal/webhook"
	"github.com/kyma-project/serverless/components/serverless/internal/webhook/resources"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	ctrlwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	scheme = runtime.NewScheme()
)

// nolint
func init() {
	_ = serverlessv1alpha2.AddToScheme(scheme)
	_ = admissionregistrationv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	setupLog := ctrlzap.New().WithName("setup")

	if !isFIPS140Only() {
		setupLog.Error(errors.New("FIPS not enforced"), "FIPS 140 exclusive mode is not enabled. Check GODEBUG flags.")
		panic("FIPS 140 exclusive mode is not enabled. Check GODEBUG flags.")
	}

	setupLog.Info("reading configuration")
	cfg := &webhook.Config{}
	if err := envconfig.InitWithPrefix(cfg, "WEBHOOK"); err != nil {
		panic(errors.Wrap(err, "while reading env variables"))
	}

	logCfg, err := logconfig.LoadConfig(cfg.LogConfigPath)
	if err != nil {
		setupLog.Error(err, "unable to load log configuration file")
		os.Exit(1)
	}

	setupLog.Info("reading webhook configuration")
	webhookCfg, err := webhook.LoadWebhookCfg(cfg.ConfigPath)
	if err != nil {
		setupLog.Error(err, "unable to load webhook configuration file")
		os.Exit(1)
	}

	atomic := zap.NewAtomicLevel()
	parsedLevel, err := zapcore.ParseLevel(logCfg.LogLevel)
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
	go logging.ReconfigureOnConfigChange(ctx, logWithCtx.Named("notifier"), atomic, cfg.ConfigPath)

	logrZap := zapr.NewLogger(logWithCtx.Desugar())
	ctrl.SetLogger(logrZap)

	// manager setup
	logWithCtx.Info("setting up controller-manager")

	mgr, err := manager.New(ctrl.GetConfigOrDie(), manager.Options{
		Scheme: scheme,
		WebhookServer: ctrlwebhook.NewServer(ctrlwebhook.Options{
			Port: cfg.Port,
		}),
		Metrics: ctrlmetrics.Options{
			BindAddress: ":9090",
		},
		Logger: logrZap,
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
		logWithCtx.Error(err, "failed to setup controller-manager")
		os.Exit(1)
	}

	logWithCtx.Info("setting up webhook certificates and webhook secret")
	// we need to ensure the certificates and the webhook secret as early as possible
	// because the webhook server needs to read it from disk to start.
	result, err := resources.SetupCertificates(context.Background(), cfg.SecretName, cfg.SystemNamespace, cfg.ServiceName,
		logWithCtx.Named("setup-certificates"))
	if err != nil {
		logWithCtx.Error(err, "failed to setup certificates and webhook secret")
		os.Exit(1)
	}
	if result == resources.Updated {
		setupLog.Info("certificate updated successfully, restarting")
		//This is not an elegant solution, but the webhook need to reconfigure itself to use updated certificate.
		//Cert-watcher from controller-runtime should refresh the certificate, but it doesn't work.
		os.Exit(0)
	}

	logWithCtx.Info("setting up webhook server")
	// webhook server setup
	whs := ctrlwebhook.NewServer(ctrlwebhook.Options{
		CertName: resources.CertFile,
		KeyName:  resources.KeyFile})
	err = whs.Start(ctx)
	if err != nil {
		logWithCtx.Error(err, "failed to start webhook server")
		os.Exit(1)
	}

	whs.Register(resources.FunctionDefaultingWebhookPath, &ctrlwebhook.Admission{
		Handler: webhook.NewDefaultingWebhook(
			mgr.GetClient(),
			logWithCtx.Named("defaulting-webhook")),
	})

	validationCfg := webhookCfg.ToValidationConfig()
	whs.Register(resources.FunctionValidationWebhookPath, &ctrlwebhook.Admission{
		Handler: webhook.NewValidatingWebhook(
			&validationCfg,
			mgr.GetClient(),
			logWithCtx.Named("validating-webhook")),
	})

	logWithCtx.Info("setting up webhook resources controller")
	// apply and monitor configuration
	if err := resources.SetupResourcesController(
		context.Background(),
		mgr,
		cfg.ServiceName,
		cfg.SystemNamespace,
		cfg.SecretName,
		logWithCtx); err != nil {
		logWithCtx.Error(err, "failed to setup webhook resources controller")
		os.Exit(1)
	}

	logWithCtx.Info("starting the controller-manager")
	// start the server manager
	err = mgr.Start(ctrl.SetupSignalHandler())
	if err != nil {
		logWithCtx.Error(err, "failed to start controller-manager")
		os.Exit(1)
	}
}

func isFIPS140Only() bool {
	return fips140.Enabled() && os.Getenv("GODEBUG") == "fips140=only,tlsmlkem=0"
}
