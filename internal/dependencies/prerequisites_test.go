package dependencies

import (
	"context"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testVirtualServiceCRD = apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: v1.ObjectMeta{
			Name: virtualServiceCRD,
		},
	}
	testGatewayCRD = apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: v1.ObjectMeta{
			Name: gatewayCRD,
		},
	}
)

func TestCheck(t *testing.T) {
	type args struct {
		ctx       context.Context
		client    client.Client
		withIstio bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "check",
			args: args{
				ctx:       context.Background(),
				client:    fake.NewFakeClientWithScheme(apiextensionsscheme.Scheme),
				withIstio: false,
			},
			wantErr: false,
		},
		{
			name: "check",
			args: args{
				ctx:       context.Background(),
				client:    fake.NewFakeClientWithScheme(apiextensionsscheme.Scheme),
				withIstio: false,
			},
			wantErr: true,
		},
		{
			name: "check with istio",
			args: args{
				ctx: context.Background(),
				client: fake.NewFakeClientWithScheme(apiextensionsscheme.Scheme,
					&testVirtualServiceCRD,
					&testGatewayCRD),
				withIstio: true,
			},
			wantErr: false,
		},
		{
			name: "check with istio fail",
			args: args{
				ctx: context.Background(),
				client: fake.NewFakeClientWithScheme(apiextensionsscheme.Scheme,
					&testGatewayCRD),
				withIstio: true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckPrerequisites(tt.args.ctx, tt.args.client, tt.args.withIstio); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
