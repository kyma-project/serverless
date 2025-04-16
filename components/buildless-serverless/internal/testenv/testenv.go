package testenv

import (
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"strings"
	"testing"
)

func Start(t *testing.T) (cl client.Client, env *envtest.Environment) {
	wdPath, err := os.Getwd()
	require.NoError(t, err)
	crdPath := buildPathFromProjectRoot(wdPath, "components", "buildless-serverless", "config", "crd", "bases")
	envtestBinsPath := buildPathFromProjectRoot(wdPath, "components", "buildless-serverless", "bin", "k8s", "kubebuilder_assets")

	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     []string{crdPath},
		ErrorIfCRDPathMissing: true,
		BinaryAssetsDirectory: envtestBinsPath,
	}
	cfg, err := testEnv.Start()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.NoError(t, scheme.AddToScheme(scheme.Scheme))
	require.NoError(t, serverlessv1alpha2.AddToScheme(scheme.Scheme))

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	require.NoError(t, err)
	require.NotNil(t, k8sClient)

	return k8sClient, testEnv
}

func Stop(t *testing.T, testEnv *envtest.Environment) {
	require.NoError(t, testEnv.Stop())
}

func buildPathFromProjectRoot(wd string, dirs ...string) string {
	wdPath := strings.Split(wd, "/")

	path := []string{"/"}
	for _, dir := range wdPath {
		if dir == "components" {
			break
		}
		path = append(path, dir)
	}
	path = append(path, dirs...)
	return filepath.Join(path...)
}
