package state

import (
	"context"
	"reflect"
	"testing"

	"github.com/kyma-project/manager-toolkit/installation/chart"
	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testStateFn = func(ctx context.Context, r *reconciler, ss *systemState) (stateFn, *ctrl.Result, error) {
		return nil, &testResult, nil
	}

	testWrappedStateFn = func(ctx context.Context, r *reconciler, ss *systemState) (stateFn, *ctrl.Result, error) {
		return testStateFn, nil, nil
	}

	testResult = ctrl.Result{
		Requeue: true,
	}

	canceledCtx = func() context.Context {
		ctx, done := context.WithCancel(context.Background())
		done()
		return ctx
	}()
)

func Test_reconciler_Reconcile(t *testing.T) {
	type fields struct {
		fn    stateFn
		log   *zap.SugaredLogger
		cache chart.ManifestCache
		k8s   k8s
		cfg   cfg
	}
	type args struct {
		ctx context.Context
		v   v1alpha1.Serverless
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ctrl.Result
		wantErr bool
	}{
		{
			name: "empty fn",
			fields: fields{
				log: zap.NewNop().Sugar(),
			},
			want:    defaultResult,
			wantErr: false,
		},
		{
			name: "with ctx done",
			fields: fields{
				log: zap.NewNop().Sugar(),
				fn:  testStateFn,
			},
			args: args{
				ctx: canceledCtx,
			},
			want:    defaultResult,
			wantErr: true,
		},
		{
			name: "with many fns",
			fields: fields{
				log: zap.NewNop().Sugar(),
				fn:  testWrappedStateFn,
			},
			args: args{
				ctx: context.Background(),
			},
			want:    testResult,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &reconciler{
				fn:    tt.fields.fn,
				log:   tt.fields.log,
				cache: tt.fields.cache,
				k8s:   tt.fields.k8s,
				cfg:   tt.fields.cfg,
			}
			got, err := m.Reconcile(tt.args.ctx, tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("reconciler.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reconciler.Reconcile() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("take status snapshot", func(t *testing.T) {
		fn := func(_ context.Context, _ *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
			// check status
			require.Equal(t, s.instance.Status, s.statusSnapshot)
			return nil, nil, nil
		}
		r := &reconciler{
			fn: fn,
			cfg: cfg{
				finalizer: v1alpha1.Finalizer,
			},
			k8s: k8s{
				client: fake.NewClientBuilder().Build(),
			},
			log: zap.NewNop().Sugar(),
		}
		serverless := v1alpha1.Serverless{
			Status: v1alpha1.ServerlessStatus{
				Conditions: []metav1.Condition{
					{
						Type:               "test-type",
						Status:             "test-status",
						Reason:             "test-reason",
						Message:            "test-message",
						ObservedGeneration: 1,
						LastTransitionTime: metav1.Now(),
					},
				},
				State: v1alpha1.StateError,
			},
		}
		_, err := r.Reconcile(context.Background(), serverless)
		require.NoError(t, err)
	})
}
