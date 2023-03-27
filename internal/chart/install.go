package chart

import (
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Install(config *Config) error {
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
		config.Log.Debugf("creating %s %s", u.GetKind(), u.GetName())
		err := config.Client.Patch(config.Ctx, &u, client.Apply, &client.PatchOptions{
			Force:        pointer.Bool(true),
			FieldManager: "keda-manager",
		})
		if err != nil {
			return err
		}
	}

	return nil
}
