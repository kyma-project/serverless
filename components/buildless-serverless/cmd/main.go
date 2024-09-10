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
	HealthzAddress      string `envconfig:"default=:8090"`
	FunctionSourcesPath string `envconfig:"default=/function-sources"`
}

func main() {
	l := logrus.New()

	cfg, err := loadConfig("APP")
	if err != nil {
		l.Error(fmt.Sprintf("unable to load config: %s", err.Error()))
		os.Exit(1)
	}

	ctrl.SetLogger(logrusr.New(l))

	l.Info("Generating Kubernetes client config")
	restConfig := ctrl.GetConfigOrDie()

	l.Info("Initializing controller manager")
	mgr, err := manager.New(restConfig, manager.Options{
		Scheme:           scheme,
		LeaderElection:   false,
		LeaderElectionID: "serverless-controller-leader-election-helper",
		// TODO: do we need this?
		// WebhookServer: ctrlwebhook.NewServer(ctrlwebhook.Options{
		// 	Port: 9443,
		// }),
		HealthProbeBindAddress: cfg.HealthzAddress,
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
		l.Error("unable to initialize controller manager", err.Error())
		os.Exit(1)
	}

	if err = mgr.AddHealthzCheck("health check", healthz.Ping); err != nil {
		l.Error("unable to register healthz", err.Error())
		os.Exit(1)
	}

	fnRecon := controllers.NewFunctionReconciler(mgr.GetClient(), l, cfg.FunctionSourcesPath)
	if err = fnRecon.SetupWithManager(mgr); err != nil {
		l.Error("unable to create Function controller", err.Error())
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	l.Info("Running manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		l.Error("unable to run the manager", err.Error())
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
