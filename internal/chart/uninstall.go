package chart

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func Uninstall(config *Config) error {
	manifest, err := getManifest(config)
	if err != nil {
		return err
	}

	objs, err := parseManifest(manifest)
	if err != nil {
		return err
	}

	for i := range objs {
		u := objs[i]
		config.Log.Debugf("deleting %s %s", u.GetKind(), u.GetName())
		err := config.Client.Delete(config.Ctx, &u)
		if errors.IsNotFound(err) {
			config.Log.Debugf("deletion skipped for %s %s", u.GetKind(), u.GetName())
			continue
		}
		if err != nil {
			return err
		}
	}

	config.Cache.Delete(types.NamespacedName{
		Name:      config.Release.Name,
		Namespace: config.Release.Namespace,
	})
	return nil
}
