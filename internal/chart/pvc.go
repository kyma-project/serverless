package chart

import (
	"context"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ MutatorFn = AdjustToClusterSize

func AdjustToClusterSize(ctx context.Context, c client.Client, obj unstructured.Unstructured) (unstructured.Unstructured, error) {
	//TODO: write test first
	return obj, nil
}
