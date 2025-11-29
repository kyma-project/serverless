package chart

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
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

	testDeployReplicaFailureCR = &appsv1.Deployment{
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
				{
					Type:    appsv1.DeploymentReplicaFailure,
					Message: "Replica failure because of test reason",
					Status:  corev1.ConditionStatus(v1.ConditionTrue),
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
		ServerlessSpecManifest{Manifest: fmt.Sprint(testCRD, separator, testDeploy)})
	_ = cache.Set(context.Background(), emptyManifestKey,
		ServerlessSpecManifest{Manifest: "---"})
	_ = cache.Set(context.Background(), wrongManifestKey,
		ServerlessSpecManifest{Manifest: "api: test\n\tversion: test"})

	type args struct {
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		want    *VerificationResult
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
			want:    &VerificationResult{Ready: true, Reason: VerificationCompleted},
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
			want:    nil,
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
			want:    &VerificationResult{Ready: true, Reason: VerificationCompleted},
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
			want:    &VerificationResult{Ready: false, Reason: DeploymentVerificationProcessing},
			wantErr: false,
		},
		{
			name: "obj replica failure",
			args: args{
				config: &Config{
					Ctx:      context.Background(),
					Log:      log,
					Cache:    cache,
					CacheKey: testManifestKey,
					Cluster: Cluster{
						Client: fake.NewClientBuilder().WithObjects(testDeployReplicaFailureCR).Build(),
					},
				},
			},
			want: &VerificationResult{
				Ready:  false,
				Reason: "deployment default/test-deploy has replica failure: Replica failure because of test reason",
			},
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
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Verify(tt.args.config)
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.wantErr, err != nil)
		})
	}
}
