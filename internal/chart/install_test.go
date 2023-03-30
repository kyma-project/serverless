package chart

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	separator = `---`
	testCRD   = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: test-crd
spec:
  group: test.group
  names:
    kind: TestKind
  versions:
    - storage: false
      name: v1alpha1
    - storage: true
      name: v1alpha2
`
	testDeploy = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deploy
  namespace: default
`
)

var (
	testDeployCR = &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-deploy",
			Namespace: "default",
		},
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionStatus(v1.ConditionTrue),
				},
			},
		},
	}
	testCRDObj = &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-crd",
		},
	}
)

func Test_install(t *testing.T) {
	log := zap.NewNop().Sugar()

	cache := NewManifestCache()
	cache.Set(types.NamespacedName{
		Name: "test", Namespace: "testnamespace",
	}, nil, fmt.Sprint(testCRD, separator, testDeploy))
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
			name: "installation error",
			args: args{
				config: &Config{
					Ctx:   context.Background(),
					Log:   log,
					Cache: cache,
					Cluster: Cluster{
						Client: fake.NewFakeClientWithScheme(apiextensionsscheme.Scheme),
					},
					Release: Release{
						Name:      "test",
						Namespace: "testnamespace",
					},
				},
			},
			wantErr: true,
		},
		// we can't simply test succeded installation here because it uses
		// tha Patch method which is not fully supported by the. This case is tested in controllers pkg
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Install(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("install() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
