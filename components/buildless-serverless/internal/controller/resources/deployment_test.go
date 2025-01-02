package resources

import (
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/config"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"testing"
)

func TestNewDeployment(t *testing.T) {
	t.Run("create deployment", func(t *testing.T) {
		f := minimalFunction()
		c := minimalFunctionConfig()

		r := NewDeployment(f, c)

		require.NotNil(t, r)
		d := r.Deployment
		require.NotNil(t, d)
		require.IsType(t, &appsv1.Deployment{}, d)
		require.Equal(t, "test-function-name", d.GetName())
		require.Equal(t, "test-function-namespace", d.GetNamespace())
	})
}

func TestDeployment_RuntimeImage(t *testing.T) {
	t.Run("return image from deployment", func(t *testing.T) {
		d := &Deployment{
			Deployment: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "test-runtime-image",
								},
							},
						},
					},
				},
			},
			functionConfig: nil,
			function:       nil,
		}

		r := d.RuntimeImage()

		require.Equal(t, "test-runtime-image", r)
	})
}

func TestDeployment_construct(t *testing.T) {
	t.Run("use runtime image from function and function config", func(t *testing.T) {
		d := minimalDeployment()

		r := d.construct()

		require.NotNil(t, r)
		require.Equal(t, r.Spec.Template.Spec.Containers[0].Image, "test-image-python312")
	})
	t.Run("use replicas from function", func(t *testing.T) {
		f := minimalFunction()
		f.Spec.Replicas = ptr.To[int32](78)
		d := &Deployment{
			Deployment:     nil,
			functionConfig: minimalFunctionConfig(),
			function:       f,
		}

		r := d.construct()

		require.NotNil(t, r)
		require.Equal(t, int32(78), *r.Spec.Replicas)
	})
	t.Run("create labels based on function", func(t *testing.T) {
		d := minimalDeployment()
		expectedLabels := map[string]string{
			"app": "test-function-name",
			"serverless.kyma-project.io/function-name": "test-function-name",
		}

		r := d.construct()

		require.NotNil(t, r)
		require.Equal(t, expectedLabels, r.ObjectMeta.Labels)
		require.Equal(t, expectedLabels, r.Spec.Selector.MatchLabels)
		require.Equal(t, expectedLabels, r.Spec.Template.ObjectMeta.Labels)
	})
	t.Run("use container name from function", func(t *testing.T) {
		d := minimalDeployment()

		r := d.construct()

		require.NotNil(t, r)
		require.Equal(t, "test-function-name", r.Spec.Template.Spec.Containers[0].Name)
	})
	t.Run("use container image based on function and function configuration", func(t *testing.T) {
		d := &Deployment{
			Deployment: nil,
			functionConfig: &config.FunctionConfig{
				ImagePython312: "special-test-image",
			},
			function: minimalFunction(),
		}

		r := d.construct()

		require.NotNil(t, r)
		require.Equal(t, "special-test-image", r.Spec.Template.Spec.Containers[0].Image)
	})
	t.Run("use container working dir based on function", func(t *testing.T) {
		d := minimalDeployment()

		r := d.construct()

		require.NotNil(t, r)
		require.Equal(t, "/kubeless", r.Spec.Template.Spec.Containers[0].WorkingDir)
	})
	t.Run("use container command dir based on function", func(t *testing.T) {
		d := minimalDeployment()

		r := d.construct()

		require.NotNil(t, r)
		require.Equal(t,
			[]string{
				"sh",
				"-c",
				`printf "${FUNC_HANDLER_SOURCE}" > handler.py;
cd ..;
python /kubeless.py;`,
			},
			r.Spec.Template.Spec.Containers[0].Command)
	})
	t.Run("use container resources based on function", func(t *testing.T) {
		rc := &serverlessv1alpha2.ResourceConfiguration{
			Function: &serverlessv1alpha2.ResourceRequirements{
				Resources: &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    k8sresource.MustParse("789m"),
						corev1.ResourceMemory: k8sresource.MustParse("678Mi"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    k8sresource.MustParse("345m"),
						corev1.ResourceMemory: k8sresource.MustParse("234Mi"),
					},
				},
			},
		}
		f := minimalFunction()
		f.Spec.ResourceConfiguration = rc
		d := &Deployment{
			Deployment:     nil,
			functionConfig: minimalFunctionConfig(),
			function:       f,
		}

		r := d.construct()

		require.NotNil(t, r)
		require.Equal(t, *rc.Function.Resources, r.Spec.Template.Spec.Containers[0].Resources)
	})
	t.Run("use container env based on function", func(t *testing.T) {
		d := minimalDeployment()
		d.function.Spec.Source.Inline.Source = "special-function-source"

		r := d.construct()

		require.NotNil(t, r)
		require.Contains(t,
			r.Spec.Template.Spec.Containers[0].Env,
			corev1.EnvVar{
				Name:  "FUNC_HANDLER_SOURCE",
				Value: "special-function-source",
			})
	})
	t.Run("use container volume mounts based on function", func(t *testing.T) {
		d := minimalDeployment()
		d.function.Spec.SecretMounts = []serverlessv1alpha2.SecretMount{
			{
				SecretName: "test-secret-name",
				MountPath:  "test-mount-path",
			},
		}

		r := d.construct()

		require.NotNil(t, r)
		require.Contains(t,
			r.Spec.Template.Spec.Containers[0].VolumeMounts,
			corev1.VolumeMount{
				Name:      "package-registry-config",
				ReadOnly:  true,
				MountPath: "/kubeless/package-registry-config/pip.conf",
				SubPath:   "pip.conf",
			})
		require.Contains(t,
			r.Spec.Template.Spec.Containers[0].VolumeMounts,
			corev1.VolumeMount{
				Name:      "test-secret-name",
				ReadOnly:  true,
				MountPath: "test-mount-path",
			})
	})
	t.Run("use volume based on function", func(t *testing.T) {
		d := minimalDeployment()
		d.function.Spec.SecretMounts = []serverlessv1alpha2.SecretMount{
			{
				SecretName: "test-secret-name",
				MountPath:  "test-mount-path",
			},
		}

		r := d.construct()

		require.NotNil(t, r)
		require.Contains(t,
			r.Spec.Template.Spec.Volumes,
			corev1.Volume{
				Name: "local",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})
		require.Contains(t,
			r.Spec.Template.Spec.Volumes,
			corev1.Volume{
				Name: "test-secret-name",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  "test-secret-name",
						DefaultMode: ptr.To[int32](0666), //read and write only for everybody
						Optional:    ptr.To[bool](false),
					},
				},
			})
	})
}

//	func TestDeployment_deploymentSecretVolumes(t *testing.T) {
//		type fields struct {
//			Deployment     *appsv1.Deployment
//			functionConfig *config.FunctionConfig
//			function       *serverlessv1alpha2.Function
//		}
//		tests := []struct {
//			name             string
//			fields           fields
//			wantVolumes      []corev1.Volume
//			wantVolumeMounts []corev1.VolumeMount
//		}{
//			// TODO: Add test cases.
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				d := &Deployment{
//					Deployment:     tt.fields.Deployment,
//					functionConfig: tt.fields.functionConfig,
//					function:       tt.fields.function,
//				}
//				gotVolumes, gotVolumeMounts := d.deploymentSecretVolumes()
//				if !reflect.DeepEqual(gotVolumes, tt.wantVolumes) {
//					t.Errorf("deploymentSecretVolumes() gotVolumes = %v, want %v", gotVolumes, tt.wantVolumes)
//				}
//				if !reflect.DeepEqual(gotVolumeMounts, tt.wantVolumeMounts) {
//					t.Errorf("deploymentSecretVolumes() gotVolumeMounts = %v, want %v", gotVolumeMounts, tt.wantVolumeMounts)
//				}
//			})
//		}
//	}
//
//	func TestDeployment_envs(t *testing.T) {
//		type fields struct {
//			Deployment     *appsv1.Deployment
//			functionConfig *config.FunctionConfig
//			function       *serverlessv1alpha2.Function
//		}
//		tests := []struct {
//			name   string
//			fields fields
//			want   []corev1.EnvVar
//		}{
//			// TODO: Add test cases.
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				d := &Deployment{
//					Deployment:     tt.fields.Deployment,
//					functionConfig: tt.fields.functionConfig,
//					function:       tt.fields.function,
//				}
//				if got := d.envs(); !reflect.DeepEqual(got, tt.want) {
//					t.Errorf("envs() = %v, want %v", got, tt.want)
//				}
//			})
//		}
//	}
//
//	func TestDeployment_name(t *testing.T) {
//		type fields struct {
//			Deployment     *appsv1.Deployment
//			functionConfig *config.FunctionConfig
//			function       *serverlessv1alpha2.Function
//		}
//		tests := []struct {
//			name   string
//			fields fields
//			want   string
//		}{
//			// TODO: Add test cases.
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				d := &Deployment{
//					Deployment:     tt.fields.Deployment,
//					functionConfig: tt.fields.functionConfig,
//					function:       tt.fields.function,
//				}
//				if got := d.name(); got != tt.want {
//					t.Errorf("name() = %v, want %v", got, tt.want)
//				}
//			})
//		}
//	}
//
//	func TestDeployment_podSpec(t *testing.T) {
//		type fields struct {
//			Deployment     *appsv1.Deployment
//			functionConfig *config.FunctionConfig
//			function       *serverlessv1alpha2.Function
//		}
//		tests := []struct {
//			name   string
//			fields fields
//			want   v1.PodSpec
//		}{
//			// TODO: Add test cases.
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				d := &Deployment{
//					Deployment:     tt.fields.Deployment,
//					functionConfig: tt.fields.functionConfig,
//					function:       tt.fields.function,
//				}
//				if got := d.podSpec(); !reflect.DeepEqual(got, tt.want) {
//					t.Errorf("podSpec() = %v, want %v", got, tt.want)
//				}
//			})
//		}
//	}
//
//	func TestDeployment_replicas(t *testing.T) {
//		type fields struct {
//			Deployment     *appsv1.Deployment
//			functionConfig *config.FunctionConfig
//			function       *serverlessv1alpha2.Function
//		}
//		tests := []struct {
//			name   string
//			fields fields
//			want   *int32
//		}{
//			// TODO: Add test cases.
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				d := &Deployment{
//					Deployment:     tt.fields.Deployment,
//					functionConfig: tt.fields.functionConfig,
//					function:       tt.fields.function,
//				}
//				if got := d.replicas(); !reflect.DeepEqual(got, tt.want) {
//					t.Errorf("replicas() = %v, want %v", got, tt.want)
//				}
//			})
//		}
//	}
//
//	func TestDeployment_resourceConfiguration(t *testing.T) {
//		type fields struct {
//			Deployment     *appsv1.Deployment
//			functionConfig *config.FunctionConfig
//			function       *serverlessv1alpha2.Function
//		}
//		tests := []struct {
//			name   string
//			fields fields
//			want   v1.ResourceRequirements
//		}{
//			// TODO: Add test cases.
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				d := &Deployment{
//					Deployment:     tt.fields.Deployment,
//					functionConfig: tt.fields.functionConfig,
//					function:       tt.fields.function,
//				}
//				if got := d.resourceConfiguration(); !reflect.DeepEqual(got, tt.want) {
//					t.Errorf("resourceConfiguration() = %v, want %v", got, tt.want)
//				}
//			})
//		}
//	}
//
//	func TestDeployment_restrictiveContainerSecurityContext(t *testing.T) {
//		type fields struct {
//			Deployment     *appsv1.Deployment
//			functionConfig *config.FunctionConfig
//			function       *serverlessv1alpha2.Function
//		}
//		tests := []struct {
//			name   string
//			fields fields
//			want   *corev1.SecurityContext
//		}{
//			// TODO: Add test cases.
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				d := &Deployment{
//					Deployment:     tt.fields.Deployment,
//					functionConfig: tt.fields.functionConfig,
//					function:       tt.fields.function,
//				}
//				if got := d.restrictiveContainerSecurityContext(); !reflect.DeepEqual(got, tt.want) {
//					t.Errorf("restrictiveContainerSecurityContext() = %v, want %v", got, tt.want)
//				}
//			})
//		}
//	}
//
//	func TestDeployment_runtimeCommand(t *testing.T) {
//		type fields struct {
//			Deployment     *appsv1.Deployment
//			functionConfig *config.FunctionConfig
//			function       *serverlessv1alpha2.Function
//		}
//		tests := []struct {
//			name   string
//			fields fields
//			want   string
//		}{
//			// TODO: Add test cases.
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				d := &Deployment{
//					Deployment:     tt.fields.Deployment,
//					functionConfig: tt.fields.functionConfig,
//					function:       tt.fields.function,
//				}
//				if got := d.runtimeCommand(); got != tt.want {
//					t.Errorf("runtimeCommand() = %v, want %v", got, tt.want)
//				}
//			})
//		}
//	}
//
//	func TestDeployment_runtimeImage(t *testing.T) {
//		type fields struct {
//			Deployment     *appsv1.Deployment
//			functionConfig *config.FunctionConfig
//			function       *serverlessv1alpha2.Function
//		}
//		tests := []struct {
//			name   string
//			fields fields
//			want   string
//		}{
//			// TODO: Add test cases.
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				d := &Deployment{
//					Deployment:     tt.fields.Deployment,
//					functionConfig: tt.fields.functionConfig,
//					function:       tt.fields.function,
//				}
//				if got := d.runtimeImage(); got != tt.want {
//					t.Errorf("runtimeImage() = %v, want %v", got, tt.want)
//				}
//			})
//		}
//	}
//
//	func TestDeployment_volumeMounts(t *testing.T) {
//		type fields struct {
//			Deployment     *appsv1.Deployment
//			functionConfig *config.FunctionConfig
//			function       *serverlessv1alpha2.Function
//		}
//		tests := []struct {
//			name   string
//			fields fields
//			want   []corev1.VolumeMount
//		}{
//			// TODO: Add test cases.
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				d := &Deployment{
//					Deployment:     tt.fields.Deployment,
//					functionConfig: tt.fields.functionConfig,
//					function:       tt.fields.function,
//				}
//				if got := d.volumeMounts(); !reflect.DeepEqual(got, tt.want) {
//					t.Errorf("volumeMounts() = %v, want %v", got, tt.want)
//				}
//			})
//		}
//	}
//
//	func TestDeployment_volumes(t *testing.T) {
//		type fields struct {
//			Deployment     *appsv1.Deployment
//			functionConfig *config.FunctionConfig
//			function       *serverlessv1alpha2.Function
//		}
//		tests := []struct {
//			name   string
//			fields fields
//			want   []corev1.Volume
//		}{
//			// TODO: Add test cases.
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				d := &Deployment{
//					Deployment:     tt.fields.Deployment,
//					functionConfig: tt.fields.functionConfig,
//					function:       tt.fields.function,
//				}
//				if got := d.volumes(); !reflect.DeepEqual(got, tt.want) {
//					t.Errorf("volumes() = %v, want %v", got, tt.want)
//				}
//			})
//		}
//	}
//
//	func TestDeployment_workingSourcesDir(t *testing.T) {
//		type fields struct {
//			Deployment     *appsv1.Deployment
//			functionConfig *config.FunctionConfig
//			function       *serverlessv1alpha2.Function
//		}
//		tests := []struct {
//			name   string
//			fields fields
//			want   string
//		}{
//			// TODO: Add test cases.
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				d := &Deployment{
//					Deployment:     tt.fields.Deployment,
//					functionConfig: tt.fields.functionConfig,
//					function:       tt.fields.function,
//				}
//				if got := d.workingSourcesDir(); got != tt.want {
//					t.Errorf("workingSourcesDir() = %v, want %v", got, tt.want)
//				}
//			})
//		}
//	}

func minimalFunction() *serverlessv1alpha2.Function {
	return &serverlessv1alpha2.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-function-name",
			Namespace: "test-function-namespace",
		},
		Spec: serverlessv1alpha2.FunctionSpec{
			Runtime: "python312",
			Source: serverlessv1alpha2.Source{
				Inline: &serverlessv1alpha2.InlineSource{
					Source: "test-function-source",
				},
			},
		},
	}
}

func minimalFunctionConfig() *config.FunctionConfig {
	return &config.FunctionConfig{
		ImagePython312: "test-image-python312",
	}
}

func minimalDeployment() *Deployment {
	return &Deployment{
		Deployment:     nil,
		functionConfig: minimalFunctionConfig(),
		function:       minimalFunction(),
	}
}
