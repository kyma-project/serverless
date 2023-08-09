package chart

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AdjustToClusterSize(ctx context.Context, c client.Client, obj unstructured.Unstructured) (unstructured.Unstructured, error) {
	//TODO: write test first
	obj.GetObjectKind()
	return obj, nil
}

func IsPVC(objKind schema.GroupVersionKind) bool {
	claim := corev1.PersistentVolumeClaim{}
	expected := claim.GroupVersionKind()

	return expected.Group == objKind.Group && expected.Kind == objKind.Kind && expected.Version == objKind.Version
}
