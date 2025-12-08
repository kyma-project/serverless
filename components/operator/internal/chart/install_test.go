package chart

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	testServiceAccount = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-service-account
  namespace: test-namespace
  labels:
    label-key: 'label-val'
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
					Status: corev1.ConditionTrue,
					Reason: MinimumReplicasAvailable,
				},
				{
					Type:   appsv1.DeploymentProgressing,
					Status: corev1.ConditionTrue,
					Reason: NewRSAvailableReason,
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

func Test_install_delete(t *testing.T) {
	t.Run("should delete all unused resources", func(t *testing.T) {
		testManifestKey := types.NamespacedName{
			Name: "test", Namespace: "testnamespace",
		}
		cache := NewInMemoryManifestCache()
		_ = cache.Set(context.Background(), testManifestKey,
			ServerlessSpecManifest{Manifest: fmt.Sprint(testCRD, separator, testDeploy)})
		client := fake.NewClientBuilder().WithObjects(testDeployCR).WithObjects(testCRDObj).Build()
		customFlags := map[string]interface{}{
			"flag1": "val1",
		}
		config := &Config{
			Cache:    cache,
			CacheKey: testManifestKey,
			Cluster: Cluster{
				Client: client,
			},
			Log: zap.NewNop().Sugar(),
		}
		err := install(config, customFlags, fixManifestRenderFunc(""))
		require.NoError(t, err)

		deploymentList := appsv1.DeploymentList{}
		err = client.List(context.Background(), &deploymentList)
		require.NoError(t, err)
		require.Empty(t, deploymentList.Items)

		crdList := apiextensionsv1.CustomResourceDefinitionList{}
		err = client.List(context.Background(), &crdList)
		require.NoError(t, err)
		require.Empty(t, crdList.Items)
	})
}

func Test_install(t *testing.T) {
	log := zap.NewNop().Sugar()

	testManifestKey := types.NamespacedName{
		Name: "test", Namespace: "testnamespace",
	}
	emptyManifestKey := types.NamespacedName{
		Name: "empty", Namespace: "manifest",
	}
	wrongManifestKey := types.NamespacedName{
		Name: "wrong", Namespace: "manifest",
	}

	cache := NewInMemoryManifestCache()
	_ = cache.Set(context.Background(), testManifestKey,
		ServerlessSpecManifest{Manifest: fmt.Sprint(testCRD, separator, testDeploy)})
	_ = cache.Set(context.Background(), emptyManifestKey,
		ServerlessSpecManifest{Manifest: ""})
	_ = cache.Set(context.Background(), wrongManifestKey,
		ServerlessSpecManifest{Manifest: "api: test\n\tversion: test"})

	type args struct {
		config      *Config
		customFlags map[string]interface{}
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
			name: "installation error",
			args: args{
				config: &Config{
					Ctx:      context.Background(),
					Log:      log,
					Cache:    cache,
					CacheKey: testManifestKey,
					Cluster: Cluster{
						Client: fake.NewClientBuilder().WithScheme(apiextensionsscheme.Scheme).Build(),
					},
				},
			},
			wantErr: true,
		},
		// we can't simply test succeded installation here because it uses
		// tha Patch method which is not fully supported by the fake client. This case is tested in controllers pkg
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Install(tt.args.config, tt.args.customFlags); (err != nil) != tt.wantErr {
				t.Errorf("install() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_unusedOldObjects(t *testing.T) {
	firstManifest := fmt.Sprint(testCRD, separator, testDeploy)
	firstObjs, _ := parseManifest(firstManifest)
	differentManifest := fmt.Sprint(testServiceAccount)
	differentObjs, _ := parseManifest(differentManifest)
	withCommonPartManifest := fmt.Sprint(testServiceAccount, separator, testDeploy)
	withCommonPartObjs, _ := parseManifest(withCommonPartManifest)
	firstWithoutCommonPartManifest := fmt.Sprint(testCRD)
	firstWithoutCommonPartObjs, _ := parseManifest(firstWithoutCommonPartManifest)

	type args struct {
		old []unstructured.Unstructured
		new []unstructured.Unstructured
	}
	tests := []struct {
		name string
		args args
		want []unstructured.Unstructured
	}{
		{
			name: "empty minus empty should be empty",
			args: args{
				old: []unstructured.Unstructured{},
				new: []unstructured.Unstructured{},
			},
			want: []unstructured.Unstructured{},
		},
		{
			name: "list minus empty should return the same list",
			args: args{
				old: firstObjs,
				new: []unstructured.Unstructured{},
			},
			want: firstObjs,
		},
		{
			name: "list minus list with different elements should return first list",
			args: args{
				old: firstObjs,
				new: differentObjs,
			},
			want: firstObjs,
		},
		{
			name: "list minus list with common part should return first list without common part",
			args: args{
				old: firstObjs,
				new: withCommonPartObjs,
			},
			want: firstWithoutCommonPartObjs,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unusedOldObjects(tt.args.old, tt.args.new)
			require.Equal(t, tt.want, got)
		})
	}
}
