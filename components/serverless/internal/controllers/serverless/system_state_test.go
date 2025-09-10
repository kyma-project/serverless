package serverless

import (
	"testing"

	"github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_systemState_podLabels(t *testing.T) {
	type args struct {
		instance *v1alpha2.Function
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Should create internal labels",
			args: args{
				instance: &v1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fn-name",
						UID:  "fn-uuid",
					},
				},
			},
			want: map[string]string{
				v1alpha2.PodAppNameLabel:        "fn-name",
				v1alpha2.FunctionUUIDLabel:      "fn-uuid",
				v1alpha2.FunctionManagedByLabel: v1alpha2.FunctionControllerValue,
				v1alpha2.FunctionNameLabel:      "fn-name",
				v1alpha2.FunctionResourceLabel:  v1alpha2.FunctionResourceLabelDeploymentValue,
			},
		},
		{
			name: "Should create internal and additional labels",
			args: args{
				instance: &v1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fn-name",
						UID:  "fn-uuid",
					},
					Spec: v1alpha2.FunctionSpec{
						Labels: map[string]string{
							"test-another": "test-another-label",
						},
					},
				},
			},
			want: map[string]string{
				v1alpha2.PodAppNameLabel:        "fn-name",
				v1alpha2.FunctionUUIDLabel:      "fn-uuid",
				v1alpha2.FunctionManagedByLabel: v1alpha2.FunctionControllerValue,
				v1alpha2.FunctionNameLabel:      "fn-name",
				v1alpha2.FunctionResourceLabel:  v1alpha2.FunctionResourceLabelDeploymentValue,
				"test-another":                  "test-another-label",
			},
		},
		{
			name: "Should create internal labels without not supported `spec.template.labels`",
			args: args{
				instance: &v1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fn-name",
						UID:  "fn-uuid",
					},
					Spec: v1alpha2.FunctionSpec{
						Template: &v1alpha2.Template{
							Labels: map[string]string{
								"test-some": "not-supported",
							},
						},
					},
				},
			},
			want: map[string]string{
				v1alpha2.PodAppNameLabel:        "fn-name",
				v1alpha2.FunctionUUIDLabel:      "fn-uuid",
				v1alpha2.FunctionManagedByLabel: v1alpha2.FunctionControllerValue,
				v1alpha2.FunctionNameLabel:      "fn-name",
				v1alpha2.FunctionResourceLabel:  v1alpha2.FunctionResourceLabelDeploymentValue,
			},
		},
		{
			name: "Should create internal and from `spec.labels` labels",
			args: args{
				instance: &v1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fn-name",
						UID:  "fn-uuid",
					},
					Spec: v1alpha2.FunctionSpec{
						Labels: map[string]string{
							"test-some": "test-label",
						},
					},
				},
			},
			want: map[string]string{
				v1alpha2.PodAppNameLabel:        "fn-name",
				v1alpha2.FunctionUUIDLabel:      "fn-uuid",
				v1alpha2.FunctionManagedByLabel: v1alpha2.FunctionControllerValue,
				v1alpha2.FunctionNameLabel:      "fn-name",
				v1alpha2.FunctionResourceLabel:  v1alpha2.FunctionResourceLabelDeploymentValue,
				"test-some":                     "test-label",
			},
		},
		{
			name: "Should not overwrite internal labels",
			args: args{
				instance: &v1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fn-name",
						UID:  "fn-uuid",
					},
					Spec: v1alpha2.FunctionSpec{
						Labels: map[string]string{
							"test-another":                 "test-label",
							v1alpha2.FunctionResourceLabel: "another-job",
							v1alpha2.FunctionNameLabel:     "another-name",
						},
					},
				},
			},
			want: map[string]string{
				v1alpha2.PodAppNameLabel:        "fn-name",
				v1alpha2.FunctionUUIDLabel:      "fn-uuid",
				v1alpha2.FunctionManagedByLabel: v1alpha2.FunctionControllerValue,
				v1alpha2.FunctionNameLabel:      "fn-name",
				v1alpha2.FunctionResourceLabel:  v1alpha2.FunctionResourceLabelDeploymentValue,
				"test-another":                  "test-label",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//GIVEN
			g := gomega.NewGomegaWithT(t)
			s := &systemState{
				instance: *tt.args.instance,
			}
			//WHEN
			got := s.podLabels()
			//THEN
			g.Expect(tt.want).To(gomega.Equal(got))
		})
	}
}

func Test_systemState_podAnnotations(t *testing.T) {
	type args struct {
		instance    *v1alpha2.Function
		deployments *appsv1.DeploymentList
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Should create internal annotations",
			args: args{
				instance: &v1alpha2.Function{},
				deployments: &appsv1.DeploymentList{
					Items: []appsv1.Deployment{},
				},
			},
			want: map[string]string{
				istioConfigLabelKey: istioEnableHoldUntilProxyStartLabelValue,
			},
		},
		{
			name: "Should create internal and from `.spec.annotations` annotations",
			args: args{
				instance: &v1alpha2.Function{
					Spec: v1alpha2.FunctionSpec{
						Annotations: map[string]string{
							"test-some": "test-annotation",
						},
					},
				},
				deployments: &appsv1.DeploymentList{
					Items: []appsv1.Deployment{},
				},
			},
			want: map[string]string{
				istioConfigLabelKey: istioEnableHoldUntilProxyStartLabelValue,
				"test-some":         "test-annotation",
			},
		},
		{
			name: "Should not overwrite internal annotations",
			args: args{
				instance: &v1alpha2.Function{
					Spec: v1alpha2.FunctionSpec{
						Annotations: map[string]string{
							"test-some":             "test-annotation",
							"proxy.istio.io/config": "another-config",
						},
					},
				},
				deployments: &appsv1.DeploymentList{
					Items: []appsv1.Deployment{},
				},
			},
			want: map[string]string{
				istioConfigLabelKey: istioEnableHoldUntilProxyStartLabelValue,
				"test-some":         "test-annotation",
			},
		},
		{
			name: "Should not overwrite internal annotations",
			args: args{
				instance: &v1alpha2.Function{
					Spec: v1alpha2.FunctionSpec{
						Annotations: map[string]string{
							"test-some":             "test-annotation",
							"proxy.istio.io/config": "another-config",
						},
					},
				},
				deployments: &appsv1.DeploymentList{
					Items: []appsv1.Deployment{},
				},
			},
			want: map[string]string{
				istioConfigLabelKey: istioEnableHoldUntilProxyStartLabelValue,
				"test-some":         "test-annotation",
			},
		},
		{
			name: "Should clear nativeSidecar annotation from deployment when not specified in function spec",
			args: args{
				instance: &v1alpha2.Function{},
				deployments: &appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						{
							Spec: appsv1.DeploymentSpec{
								Template: corev1.PodTemplateSpec{
									ObjectMeta: metav1.ObjectMeta{
										Annotations: map[string]string{
											"sidecar.istio.io/nativeSidecar": "true",
										},
									},
								},
							},
						},
					},
				},
			},
			want: map[string]string{
				istioConfigLabelKey: istioEnableHoldUntilProxyStartLabelValue,
			},
		},
		{
			name: "Should keep nativeSidecar annotation in deployment when specified in function spec",
			args: args{
				instance: &v1alpha2.Function{
					Spec: v1alpha2.FunctionSpec{
						Annotations: map[string]string{
							"sidecar.istio.io/nativeSidecar": "true",
						},
					},
				},
				deployments: &appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						{
							Spec: appsv1.DeploymentSpec{
								Template: corev1.PodTemplateSpec{
									ObjectMeta: metav1.ObjectMeta{
										Annotations: map[string]string{
											"sidecar.istio.io/nativeSidecar": "true",
										},
									},
								},
							},
						},
					},
				},
			},
			want: map[string]string{
				istioConfigLabelKey:              istioEnableHoldUntilProxyStartLabelValue,
				"sidecar.istio.io/nativeSidecar": "true",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//GIVEN
			g := gomega.NewGomegaWithT(t)
			s := &systemState{
				instance:    *tt.args.instance,
				deployments: *tt.args.deployments,
			}
			//WHEN
			got := s.podAnnotations()
			//THEN
			g.Expect(tt.want).To(gomega.Equal(got))
		})
	}
}
