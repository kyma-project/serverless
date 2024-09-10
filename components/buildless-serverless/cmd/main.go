package main

import (
	"fmt"
	"os"

	"github.com/bombsimon/logrusr/v4"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controllers"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/pkg/apis/serverless/v1alpha2"
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	ctrlwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
	// +kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

// nolint
func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = serverlessv1alpha2.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type config struct {
	MetricsAddress            string `envconfig:"default=:8080"`
	Healthz                   healthzConfig
	LeaderElectionEnabled     bool   `envconfig:"default=false"`
	LeaderElectionID          string `envconfig:"default=serverless-controller-leader-election-helper"`
	SecretMutatingWebhookPort int    `envconfig:"default=8443"`
	// Function                  serverless.FunctionConfig
}

type healthzConfig struct {
	Address string `envconfig:"default=:8090"`
}

func main() {
	l := logrus.New()

	config, err := loadConfig("APP")
	if err != nil {
		l.Error(fmt.Sprintf("unable to load config: %s", err.Error()))
		os.Exit(1)
	}

	ctrl.SetLogger(logrusr.New(l))

	l.Info("Generating Kubernetes client config")
	restConfig := ctrl.GetConfigOrDie()

	l.Info("Initializing controller manager")
	mgr, err := manager.New(restConfig, manager.Options{
		Scheme: scheme,
		Metrics: ctrlmetrics.Options{
			BindAddress: config.MetricsAddress,
		},
		LeaderElection:   config.LeaderElectionEnabled,
		LeaderElectionID: config.LeaderElectionID,
		WebhookServer: ctrlwebhook.NewServer(ctrlwebhook.Options{
			Port: config.SecretMutatingWebhookPort,
		}),
		HealthProbeBindAddress: config.Healthz.Address,
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
		l.Error(fmt.Sprintf("unable to initialize controller manager", err.Error()))
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("health check", healthz.Ping); err != nil {
		l.Error(fmt.Sprintf("unable to register healthz", err.Error()))
		os.Exit(1)
	}

	fnRecon := controllers.NewFunctionReconciler(mgr.GetClient(), l)
	if err != fnRecon.SetupWithManager(mgr) {
		l.Error(fmt.Sprintf("unable to create Function controller", err.Error()))
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	l.Info("Running manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		l.Error(fmt.Sprintf("unable to run the manager", err.Error()))
		os.Exit(1)
	}
}

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
