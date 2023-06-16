package chart

import (
	"fmt"

	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO: cover case when user change enableInternal

func Install(config *Config) error {
	manifest, err := getOrRenderManifest(config)
	if err != nil {
		return fmt.Errorf("could not render manifest from chart: %s", err.Error())
	}

	objs, err := parseManifest(manifest)
	if err != nil {
		return fmt.Errorf("could not parse chart manifest: %s", err.Error())
	}

	for i := range objs {
		u := objs[i]
		config.Log.Debugf("creating %s %s/%s", u.GetKind(), u.GetNamespace(), u.GetName())

		// TODO: what if Path returns error in the middle of manifest?
		// maybe we should in this case translate applied objs into manifest and set it into cache?
		err := config.Cluster.Client.Patch(config.Ctx, &u, client.Apply, &client.PatchOptions{
			Force:        pointer.Bool(true),
			FieldManager: "serverless-manager",
		})
		if err != nil {
			return fmt.Errorf("could not install object %s/%s: %s", u.GetNamespace(), u.GetName(), err.Error())
		}
	}

	return config.Cache.Set(config.Ctx, config.CacheKey, ServerlessSpecManifest{
		ManagerUID:  config.ManagerUID,
		CustomFlags: config.Release.Flags,
		Manifest:    manifest,
	})
}
