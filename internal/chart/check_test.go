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
	cache := NewManifestCache()
	cache.Set(types.NamespacedName{
		Name: "no", Namespace: "crd",
	}, nil, fmt.Sprint(testDeploy))
	cache.Set(types.NamespacedName{
		Name: "no", Namespace: "orphan",
	}, nil, fmt.Sprint(testCRD, separator, testDeploy))
	cache.Set(types.NamespacedName{
		Name: "one", Namespace: "orphan",
	}, nil, fmt.Sprint(testCRD, separator, testOrphanCR))
	cache.Set(types.NamespacedName{
		Name: "empty", Namespace: "manifest",
	}, nil, "")
	cache.Set(types.NamespacedName{
		Name: "wrong", Namespace: "manifest",
	}, nil, "api: test\n\tversion: test")

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
					Cache: cache,
					Release: Release{
						Name:      "empty",
						Namespace: "manifest",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "parse manifest error",
			args: args{
				config: &Config{
					Cache: cache,
					Release: Release{
						Name:      "wrong",
						Namespace: "manifest",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "no CRDs in manifest",
			args: args{
				config: &Config{
					Cache: cache,
					Release: Release{
						Name:      "no",
						Namespace: "crd",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "no orphan for CRD",
			args: args{
				config: &Config{
					Cache: cache,
					Ctx:   context.Background(),
					Cluster: Cluster{
						Client: fake.NewFakeClientWithScheme(apiextensionsscheme.Scheme, testCRDObj),
					},
					Release: Release{
						Name:      "no",
						Namespace: "orphan",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "one orphan for CRD",
			args: args{
				config: &Config{
					Cache: cache,
					Ctx:   context.Background(),
					Cluster: Cluster{
						Client: func() client.Client {
							scheme := runtime.NewScheme()
							scheme.AddKnownTypes(schema.GroupVersion{
								Group:   "test.group",
								Version: "v1alpha2",
							}, &testOrphanObj)
							apiextensionsscheme.AddToScheme(scheme)
							c := fake.NewFakeClientWithScheme(scheme, &testOrphanObj, testCRDObj)

							return c
						}(),
					},
					Release: Release{
						Name:      "one",
						Namespace: "orphan",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing CRD on cluster",
			args: args{
				config: &Config{
					Cache: cache,
					Ctx:   context.Background(),
					Cluster: Cluster{
						Client: func() client.Client {
							scheme := runtime.NewScheme()
							apiextensionsscheme.AddToScheme(scheme)
							c := fake.NewFakeClientWithScheme(scheme)

							return c
						}(),
					},
					Release: Release{
						Name:      "one",
						Namespace: "orphan",
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
