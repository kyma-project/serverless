package chart

import (
	"fmt"

	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Install(config *Config) error {
	manifest, err := getManifest(config)
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
		err := config.Client.Patch(config.Ctx, &u, client.Apply, &client.PatchOptions{
			Force:        pointer.Bool(true),
			FieldManager: "keda-manager",
		})
		if err != nil {
			return fmt.Errorf("could not install object %s/%s: %s", u.GetNamespace(), u.GetName(), err.Error())
		}
	}

	return nil
}
