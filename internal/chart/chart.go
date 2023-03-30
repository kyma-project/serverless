package chart

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	Ctx     context.Context
	Log     *zap.SugaredLogger
	Cache   *ManifestCache
	Cluster Cluster
	Release Release
}

type Release struct {
	Flags     map[string]interface{}
	ChartPath string
	Name      string
	Namespace string
}

type Cluster struct {
	Client client.Client
	Config *rest.Config
}

func parseManifest(manifest string) ([]unstructured.Unstructured, error) {
	results := make([]unstructured.Unstructured, 0)
	decoder := yaml.NewDecoder(strings.NewReader(manifest))

	for {
		var obj map[string]interface{}
		err := decoder.Decode(&obj)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		// no obj between separators
		if len(obj) == 0 {
			continue
		}

		u := unstructured.Unstructured{Object: obj}
		if u.GetObjectKind().GroupVersionKind().Kind == "CustomResourceDefinition" {
			results = append([]unstructured.Unstructured{u}, results...)
			continue
		}
		results = append(results, u)
	}

	return results, nil
}

func getManifest(config *Config) (string, error) {
	cacheKey := types.NamespacedName{
		Name:      config.Release.Name,
		Namespace: config.Release.Namespace,
	}

	specManifest := config.Cache.Get(cacheKey)
	if specManifest != nil && reflect.DeepEqual(specManifest.customFlags, config.Release.Flags) {
		return specManifest.manifest, nil
	}

	release, err := renderChart(config)
	if err != nil {
		return "", err
	}

	config.Cache.Set(cacheKey, config.Release.Flags, release.Manifest)
	return release.Manifest, nil
}

func renderChart(config *Config) (*release.Release, error) {
	chart, err := loader.Load(config.Release.ChartPath)
	if err != nil {
		return nil, fmt.Errorf("while loading chart from path '%s': %s", config.Release.ChartPath, err.Error())
	}

	installAction := newInstallAction(config)

	rel, err := installAction.Run(chart, config.Release.Flags)
	if err != nil {
		return nil, fmt.Errorf("while templating chart: %s", err.Error())
	}

	return rel, nil
}

func newInstallAction(config *Config) *action.Install {
	helmRESTGetter := &clientGetter{
		config: config.Cluster.Config,
	}

	helmClient := kube.New(helmRESTGetter)
	helmClient.Log = config.Log.Debugf

	actionConfig := new(action.Configuration)
	actionConfig.KubeClient = helmClient
	actionConfig.Log = helmClient.Log

	actionConfig.Releases = storage.Init(driver.NewMemory())
	actionConfig.RESTClientGetter = helmRESTGetter

	action := action.NewInstall(actionConfig)
	action.ReleaseName = config.Release.Name
	action.Namespace = config.Release.Namespace
	action.Replace = true
	action.IsUpgrade = true
	action.DryRun = true

	return action
}
