package chart

import (
	"context"
	"testing"

	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/types"
)

func Test_getOrRenderManifestWithRenderer(t *testing.T) {
	noCRDManifestKey := types.NamespacedName{
		Name: "no", Namespace: "crd",
	}

	cache := NewInMemoryManifestCache()
	_ = cache.Set(context.Background(), noCRDManifestKey,
		DockerRegistrySpecManifest{Manifest: testDeploy})

	type args struct {
		config          *Config
		customFlags     map[string]interface{}
		renderChartFunc func(config *Config, customFlags map[string]interface{}) (*release.Release, error)
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "return manifest when flags and managerUID are not changed",
			args: args{
				config: &Config{
					Ctx:      context.Background(),
					Cache:    cache,
					CacheKey: noCRDManifestKey,
				},
			},
			want:    testDeploy,
			wantErr: false,
		},
		{
			name: "render manifest when flags are changed",
			args: args{
				renderChartFunc: fixManifestRenderFunc("test-new-manifest"),
				customFlags: map[string]interface{}{
					"flag1": "val1",
				},
				config: &Config{
					Ctx:      context.Background(),
					Cache:    cache,
					CacheKey: noCRDManifestKey,
				},
			},
			want:    "test-new-manifest",
			wantErr: false,
		},
		{
			name: "render manifest when managerUID is changed",
			args: args{
				renderChartFunc: fixManifestRenderFunc("test-new-manifest-2"),
				config: &Config{
					Ctx:        context.Background(),
					Cache:      cache,
					CacheKey:   noCRDManifestKey,
					ManagerUID: "new-UID",
				},
			},
			want:    "test-new-manifest-2",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotCurrent, err := getCachedAndCurrentManifest(tt.args.config, tt.args.customFlags, tt.args.renderChartFunc)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCachedAndCurrentManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCurrent != tt.want {
				t.Errorf("getCachedAndCurrentManifest() = %v, want %v", gotCurrent, tt.want)
			}
		})
	}
}

func fixManifestRenderFunc(manifest string) func(config *Config, customFlags map[string]interface{}) (*release.Release, error) {
	return func(config *Config, customFlags map[string]interface{}) (*release.Release, error) {
		return &release.Release{
			Manifest: manifest,
		}, nil
	}
}
