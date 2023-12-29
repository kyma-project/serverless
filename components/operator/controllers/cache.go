package controllers

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CacheCreator(cfg *rest.Config, opts cache.Options) (cache.Cache, error) {
	labelSelector, err := labels.Parse("app.kubernetes.io/part-of in (serverless,serverless)")
	if err != nil {
		panic(fmt.Sprintf("unable to parse label selector: %s", err))
	}
	objSelector := cache.ObjectSelector{
		Label: labelSelector,
	}
	opts.SelectorsByObject = map[client.Object]cache.ObjectSelector{
		&corev1.Secret{}:    objSelector,
		&corev1.ConfigMap{}: objSelector,
	}
	return cache.New(cfg, opts)
}
