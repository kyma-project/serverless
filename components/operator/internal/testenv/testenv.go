package testenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1alpha1 "github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func Start(t *testing.T) (cl client.Client, env *envtest.Environment) {
	wdPath, err := os.Getwd()
	require.NoError(t, err)
	crdPath := buildCrdPath(wdPath)

	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     []string{crdPath},
		ErrorIfCRDPathMissing: true,
	}
	cfg, err := testEnv.Start()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.NoError(t, scheme.AddToScheme(scheme.Scheme))
	require.NoError(t, v1alpha1.AddToScheme(scheme.Scheme))

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	require.NoError(t, err)
	require.NotNil(t, k8sClient)

	return k8sClient, testEnv
}

func Stop(t *testing.T, testEnv *envtest.Environment) {
	require.NoError(t, testEnv.Stop())
}

func buildCrdPath(wd string) string {
	wdPath := strings.Split(wd, "/")

	crdPath := []string{"/"}
	for _, path := range wdPath {
		crdPath = append(crdPath, path)
		if path == "components" {
			break
		}
	}
	crdPath = append(crdPath, "..", "config", "operator", "base", "crd", "bases")
	return filepath.Join(crdPath...)
}
