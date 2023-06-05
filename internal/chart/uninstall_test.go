package chart

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/zap"
	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_Uninstall(t *testing.T) {
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
	cache.Set(context.Background(), testManifestKey, nil, fmt.Sprint(testCRD, separator, testDeploy))
	cache.Set(context.Background(), emptyManifestKey, nil, "")
	cache.Set(context.Background(), wrongManifestKey, nil, "api: test\n\tversion: test")

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
			name: "installation error",
			args: args{
				config: &Config{
					Ctx:      context.Background(),
					Log:      log,
					Cache:    cache,
					CacheKey: testManifestKey,
					Cluster: Cluster{
						Client: fake.NewFakeClientWithScheme(apiextensionsscheme.Scheme),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Uninstall(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("uninstall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
