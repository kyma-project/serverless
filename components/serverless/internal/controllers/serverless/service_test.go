package serverless

import (
	"context"
	"github.com/kyma-project/serverless/components/serverless/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"

	"go.uber.org/zap"

	"github.com/kyma-project/serverless/components/serverless/internal/resource/automock"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
)

func TestFunctionReconciler_equalServices(t *testing.T) {
	type args struct {
		existing corev1.Service
		expected corev1.Service
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "simple equal case",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "svc-name",
						Namespace:   "svc-ns",
						Annotations: prometheusSvcAnnotations(),
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
						Annotations: prometheusSvcAnnotations(),
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "fails on different labels",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"different": "label-different",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "fails on different port",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       8000,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "fails on different port name",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "httpzzzz",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "fails on different targetPort",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(666)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "fails on different selector",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value-DIFFERENT",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "fails if there is 0 ports in existing",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "fails if there is 0 ports in expected",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name: "test",
						}},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "fails if there is 0 ports in either case",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			got := equalServices(tt.args.existing, tt.args.expected)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_deleteExcessServices(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)

		instance := &serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{Name: "fn-name"},
		}

		services := []corev1.Service{
			{ObjectMeta: metav1.ObjectMeta{Name: "fn-name"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "fn-some-other-name"}},
		}

		client := new(automock.Client)
		client.On("Delete", context.TODO(), &services[1]).Return(nil).Once()
		defer client.AssertExpectations(t)

		s := systemState{
			instance: *instance,
			services: corev1.ServiceList{
				Items: services,
			},
		}

		r := reconciler{
			log: zap.NewNop().Sugar(),
			k8s: k8s{
				client: client,
			},
		}

		_, err := stateFnDeleteServices(context.TODO(), &r, &s)

		g.Expect(err).To(gomega.Succeed())
		g.Expect(client.Calls).To(gomega.HaveLen(1), "delete should happen only for service which has different name than it's parent fn")
	})

	t.Run("should delete both svc that have different name than fn", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)

		instance := &serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{Name: "fn-name"},
		}

		services := []corev1.Service{
			{ObjectMeta: metav1.ObjectMeta{Name: "fn-other-name"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "fn-some-other-name"}},
		}

		client := new(automock.Client)
		client.On("Delete", context.TODO(), &services[0]).Return(nil).Once()
		client.On("Delete", context.TODO(), &services[1]).Return(nil).Once()
		defer client.AssertExpectations(t)

		s := systemState{
			instance: *instance,
			services: corev1.ServiceList{
				Items: services,
			},
		}

		r := reconciler{
			log: zap.NewNop().Sugar(),
			k8s: k8s{
				client: client,
			},
		}

		_, err := stateFnDeleteServices(context.TODO(), &r, &s)

		g.Expect(err).To(gomega.Succeed())
		g.Expect(client.Calls).To(gomega.HaveLen(2), "delete should happen only for service which has different name than it's parent fn")
	})
}

func TestFunctionReconciler_buildStateFnUpdateService(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	oldAnnotationKey := "old"
	updatedAnnotationKey := "updated"
	newAnnotationKey := "new"

	testCases := map[string]struct {
		oldSvcAnnotations      map[string]string
		newSvcAnnotations      map[string]string
		expectedSvcAnnotations map[string]string
	}{
		"Svc is updated with merged annotations": {
			oldSvcAnnotations: map[string]string{
				oldAnnotationKey:     "old-value",
				updatedAnnotationKey: "old-value-to-update",
			},
			newSvcAnnotations: map[string]string{
				updatedAnnotationKey: "updated-value",
				newAnnotationKey:     "new-value",
			},
			expectedSvcAnnotations: map[string]string{
				oldAnnotationKey:     "old-value",
				updatedAnnotationKey: "updated-value",
				newAnnotationKey:     "new-value",
			},
		},
		"Svc is updated when old svc annotations are empty": {
			newSvcAnnotations: map[string]string{
				updatedAnnotationKey: "updated-value",
				newAnnotationKey:     "new-value",
			},
			expectedSvcAnnotations: map[string]string{
				updatedAnnotationKey: "updated-value",
				newAnnotationKey:     "new-value",
			},
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			oldSvc, newSvc := fixSvc()
			oldSvc.Annotations = testData.oldSvcAnnotations
			newSvc.Annotations = testData.newSvcAnnotations

			builder := &fake.ClientBuilder{}
			k8sClient := resource.New(builder.WithObjects(&oldSvc).Build(), scheme.Scheme)
			r := &reconciler{k8s: k8s{client: k8sClient}, log: zap.NewNop().Sugar()}
			s := &systemState{services: corev1.ServiceList{Items: []corev1.Service{oldSvc}}}
			//WHEN
			fn := buildStateFnUpdateService(newSvc)
			_, err := fn(ctx, r, s)

			//THEN
			require.NoError(t, err)
			updatedSvc := corev1.Service{}
			require.NoError(t, k8sClient.Get(ctx, client.ObjectKey{Namespace: oldSvc.GetNamespace(), Name: oldSvc.GetName()}, &updatedSvc))
			assert.Equal(t, newSvc.Spec.Ports, updatedSvc.Spec.Ports)
			assert.Equal(t, newSvc.Spec.Selector, updatedSvc.Spec.Selector)
			assert.Equal(t, newSvc.Spec.Type, updatedSvc.Spec.Type)
			assert.EqualValues(t, newSvc.GetLabels(), updatedSvc.GetLabels())
			assert.EqualValues(t, testData.expectedSvcAnnotations, updatedSvc.GetAnnotations())
		})
	}

}

func fixSvc() (old corev1.Service, new corev1.Service) {

	oldSvc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "test-service",
			Labels: map[string]string{
				"A": "B",
				"C": "D",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "old-port", Port: int32(1234)},
				{Name: "another-port", Port: int32(4321)},
			},
			Selector: map[string]string{
				"selA": "B",
				"selC": "D",
			},
			Type: corev1.ServiceTypeClusterIP},
	}

	newSvc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "test-service",
			Labels: map[string]string{
				"X": "Y",
				"W": "Z",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "new-port", Port: int32(9876)},
				{Name: "another=new-port", Port: int32(6789)},
			},
			Selector: map[string]string{
				"selX": "Y",
				"selW": "Z",
			},
			Type: corev1.ServiceTypeLoadBalancer},
	}
	return oldSvc, newSvc
}
