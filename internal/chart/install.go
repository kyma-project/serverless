package chart

import (
	"fmt"
	"github.com/pkg/errors"

	"github.com/kyma-project/serverless-manager/internal/annotation"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Install(config *Config) error {
	return install(config, renderChart)
}

func install(config *Config, renderChartFunc func(config *Config) (*release.Release, error)) error {
	cachedManifest, currentManifest, err := getCachedAndCurrentManifest(config, renderChartFunc)
	if err != nil {
		return err
	}

	objs, unusedObjs, err := getObjectsToInstallAndRemove(cachedManifest, currentManifest)
	if err != nil {
		return err
	}

	err = updateObjects(config, objs)
	if err != nil {
		return err
	}

	uninstallObjects(config, unusedObjs)

	return config.Cache.Set(config.Ctx, config.CacheKey, ServerlessSpecManifest{
		ManagerUID:  config.ManagerUID,
		CustomFlags: config.Release.Flags,
		Manifest:    currentManifest,
	})
}

func getObjectsToInstallAndRemove(cachedManifest string, currentManifest string) ([]unstructured.Unstructured, []unstructured.Unstructured, error) {
	objs, err := parseManifest(currentManifest)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse chart manifest: %s", err.Error())
	}

	oldObjs, err := parseManifest(cachedManifest)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse chart manifest: %s", err.Error())
	}

	unusedObjs := unusedOldObjects(oldObjs, objs)
	return objs, unusedObjs, nil
}

func updateObjects(config *Config, objs []unstructured.Unstructured) error {
	for i := range objs {
		u := objs[i]
		config.Log.Debugf("creating %s %s/%s", u.GetKind(), u.GetNamespace(), u.GetName())

		u = annotation.AddDoNotEditDisclaimer(u)
		if IsPVC(u.GroupVersionKind()) {
			modifiedObj, err := AdjustToClusterSize(config.Ctx, config.Cluster.Client, u)
			if err != nil {
				return errors.Wrap(err, "while adjusting pvc size")
			}
			u = modifiedObj
		}

		// TODO: what if Path returns error in the middle of manifest?
		// maybe we should in this case translate applied objs into manifest and set it into cache?
		err := config.Cluster.Client.Patch(config.Ctx, &u, client.Apply, &client.PatchOptions{
			Force:        pointer.Bool(true),
			FieldManager: "serverless-operator",
		})
		if err != nil {
			return fmt.Errorf("could not install object %s/%s: %s", u.GetNamespace(), u.GetName(), err.Error())
		}

		// remove old reconciler "DO NOT EDIT" disclaimer
		// TODO: remove this functionality when all clusters are migrated to the serverless-manager
		err = annotation.DeleteReconcilerDisclaimer(
			config.Cluster.Client, *config.Cluster.Config, u)
		if err != nil {
			return fmt.Errorf("could not remove old reconciler annotation for object %s/%s: %s",
				u.GetNamespace(), u.GetName(), err.Error())
		}
	}
	return nil
}

func unusedOldObjects(previousObjs []unstructured.Unstructured, currentObjs []unstructured.Unstructured) []unstructured.Unstructured {
	currentNames := make(map[string]struct{}, len(currentObjs))
	for _, obj := range currentObjs {
		objFullName := fmt.Sprintf("%s/%s/%s", obj.GetKind(), obj.GetNamespace(), obj.GetName())
		currentNames[objFullName] = struct{}{}
	}
	result := []unstructured.Unstructured{}
	for _, obj := range previousObjs {
		objFullName := fmt.Sprintf("%s/%s/%s", obj.GetKind(), obj.GetNamespace(), obj.GetName())
		if _, found := currentNames[objFullName]; !found {
			result = append(result, obj)
		}
	}
	return result
}
