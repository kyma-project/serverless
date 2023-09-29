package predicate

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestNoStatusChangePredicate_Update(t *testing.T) {
	type args struct {
		e event.UpdateEvent
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil objs",
			args: args{
				e: event.UpdateEvent{
					ObjectOld: nil,
					ObjectNew: nil,
				},
			},
			want: false,
		},
		{
			name: "first obj iteration",
			args: args{
				e: event.UpdateEvent{
					ObjectOld: func() *unstructured.Unstructured {
						u := &unstructured.Unstructured{}
						u.SetGeneration(1)
						u.SetResourceVersion("560")
						return u
					}(),
					ObjectNew: func() *unstructured.Unstructured {
						u := &unstructured.Unstructured{}
						u.SetGeneration(1)
						u.SetResourceVersion("560")
						return u
					}(),
				},
			},
			want: true,
		},
		{
			name: "status update",
			args: args{
				e: event.UpdateEvent{
					ObjectOld: func() *unstructured.Unstructured {
						u := &unstructured.Unstructured{}
						u.SetGeneration(1)
						u.SetResourceVersion("560")
						return u
					}(),
					ObjectNew: func() *unstructured.Unstructured {
						u := &unstructured.Unstructured{}
						u.SetGeneration(1)
						u.SetResourceVersion("600")
						return u
					}(),
				},
			},
			want: false,
		},
		{
			name: "spec update",
			args: args{
				e: event.UpdateEvent{
					ObjectOld: func() *unstructured.Unstructured {
						u := &unstructured.Unstructured{}
						u.SetGeneration(1)
						u.SetResourceVersion("560")
						return u
					}(),
					ObjectNew: func() *unstructured.Unstructured {
						u := &unstructured.Unstructured{}
						u.SetGeneration(2)
						u.SetResourceVersion("600")
						return u
					}(),
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NoStatusChangePredicate{}
			if got := p.Update(tt.args.e); got != tt.want {
				t.Errorf("NoStatusChangePredicate.Update() = %v, want %v", got, tt.want)
			}
		})
	}
}
