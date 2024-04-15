package chart

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testDeployNotReadyCR = &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-deploy",
			Namespace: "default",
		},
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionStatus(v1.ConditionFalse),
				},
			},
		},
	}
)

func Test_verify(t *testing.T) {
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
		DockerRegistrySpecManifest{Manifest: fmt.Sprint(testCRD, separator, testDeploy)})
	_ = cache.Set(context.Background(), emptyManifestKey,
		DockerRegistrySpecManifest{Manifest: "---"})
	_ = cache.Set(context.Background(), wrongManifestKey,
		DockerRegistrySpecManifest{Manifest: "api: test\n\tversion: test"})

	type args struct {
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		want    bool
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
			want:    true,
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
			want:    false,
			wantErr: true,
		},
		{
			name: "verify",
			args: args{
				config: &Config{
					Ctx:      context.Background(),
					Log:      log,
					Cache:    cache,
					CacheKey: testManifestKey,
					Cluster: Cluster{
						Client: fake.NewClientBuilder().WithObjects(testDeployCR).Build(),
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "obj not ready",
			args: args{
				config: &Config{
					Ctx:      context.Background(),
					Log:      log,
					Cache:    cache,
					CacheKey: testManifestKey,
					Cluster: Cluster{
						Client: fake.NewClientBuilder().WithObjects(testDeployNotReadyCR).Build(),
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "obj not found",
			args: args{
				config: &Config{
					Ctx:      context.Background(),
					Log:      log,
					Cache:    cache,
					CacheKey: testManifestKey,
					Cluster: Cluster{
						Client: fake.NewClientBuilder().Build(),
					},
				},
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Verify(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("verify() = %v, want %v", got, tt.want)
			}
		})
	}
}
