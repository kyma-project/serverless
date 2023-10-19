package chart

import (
	"context"
	"fmt"
	"testing"

	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testOrphanCR = `
apiVersion: test.group/v1alpha2
kind: TestKind
metadata:
  name: test-deploy
  namespace: default
`
)

var (
	testOrphanObj = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "test.group/v1alpha2",
			"kind":       "TestKind",
			"metadata": map[string]interface{}{
				"name":      "test",
				"namespace": "namespace",
			},
		},
	}
)

func TestCheckCRDOrphanResources(t *testing.T) {
	noCRDManifestKey := types.NamespacedName{
		Name: "no", Namespace: "crd",
	}
	noOrphanManifestKey := types.NamespacedName{
		Name: "no", Namespace: "orphan",
	}
	oneOrphanManifestKey := types.NamespacedName{
		Name: "one", Namespace: "orphan",
	}
	emptyManifestKey := types.NamespacedName{
		Name: "empty", Namespace: "manifest",
	}
	wrongManifestKey := types.NamespacedName{
		Name: "wrong", Namespace: "manifest",
	}

	cache := NewInMemoryManifestCache()
	cache.Set(context.Background(), noCRDManifestKey,
		ServerlessSpecManifest{Manifest: fmt.Sprint(testDeploy)})
	cache.Set(context.Background(), noOrphanManifestKey,
		ServerlessSpecManifest{Manifest: fmt.Sprint(testCRD, separator, testDeploy)})
	cache.Set(context.Background(), oneOrphanManifestKey,
		ServerlessSpecManifest{Manifest: fmt.Sprint(testCRD, separator, testOrphanCR)})
	cache.Set(context.Background(), emptyManifestKey,
		ServerlessSpecManifest{Manifest: ""})
	cache.Set(context.Background(), wrongManifestKey,
		ServerlessSpecManifest{Manifest: "api: test\n\tversion: test"})

	type args struct {
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty manifest",
			args: args{
				config: &Config{
					Cache:    cache,
					CacheKey: emptyManifestKey,
				},
			},
			wantErr: false,
		},
		{
			name: "parse manifest error",
			args: args{
				config: &Config{
					Cache:    cache,
					CacheKey: wrongManifestKey,
				},
			},
			wantErr: true,
		},
		{
			name: "no CRDs in manifest",
			args: args{
				config: &Config{
					Cache:    cache,
					CacheKey: noCRDManifestKey,
				},
			},
			wantErr: false,
		},
		{
			name: "no orphan for CRD",
			args: args{
				config: &Config{
					Cache:    cache,
					CacheKey: noOrphanManifestKey,
					Ctx:      context.Background(),
					Cluster: Cluster{
						Client: fake.NewClientBuilder().
							WithScheme(apiextensionsscheme.Scheme).
							WithObjects(testCRDObj).
							Build(),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "one orphan for CRD",
			args: args{
				config: &Config{
					Cache:    cache,
					CacheKey: oneOrphanManifestKey,
					Ctx:      context.Background(),
					Cluster: Cluster{
						Client: func() client.Client {
							scheme := runtime.NewScheme()
							scheme.AddKnownTypes(schema.GroupVersion{
								Group:   "test.group",
								Version: "v1alpha2",
							}, &testOrphanObj)
							apiextensionsscheme.AddToScheme(scheme)
							c := fake.NewClientBuilder().
								WithScheme(scheme).
								WithObjects(&testOrphanObj).
								WithObjects(testCRDObj).
								Build()
							return c
						}(),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing CRD on cluster",
			args: args{
				config: &Config{
					Cache:    cache,
					CacheKey: oneOrphanManifestKey,
					Ctx:      context.Background(),
					Cluster: Cluster{
						Client: func() client.Client {
							scheme := runtime.NewScheme()
							apiextensionsscheme.AddToScheme(scheme)
							c := fake.NewClientBuilder().WithScheme(scheme).Build()
							return c
						}(),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckCRDOrphanResources(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("CheckCRDOrphanResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
