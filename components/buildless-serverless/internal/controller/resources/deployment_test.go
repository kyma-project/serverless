package resources

import (
	"testing"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestNewDeployment(t *testing.T) {
	t.Run("create deployment", func(t *testing.T) {
		f := minimalFunction()
		c := minimalFunctionConfig()

		r := NewDeployment(f, c, nil, "test-commit", nil)

		require.NotNil(t, r)
		d := r.Deployment
		require.NotNil(t, d)
		require.IsType(t, &appsv1.Deployment{}, d)
		require.Equal(t, "test-function-name-", d.GetGenerateName())
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
		d.function.Spec.Labels = map[string]string{
			"shtern": "stoic",
			"boyd":   "vigilant",
		}

		r := d.construct()

		require.NotNil(t, r)
		require.Equal(t, map[string]string{
			"serverless.kyma-project.io/function-name": "test-function-name",
			"serverless.kyma-project.io/managed-by":    "function-controller",
			"serverless.kyma-project.io/uuid":          "test-uid",
		}, r.ObjectMeta.Labels)
		require.Equal(t, map[string]string{
			"serverless.kyma-project.io/function-name": "test-function-name",
			"serverless.kyma-project.io/managed-by":    "function-controller",
			"serverless.kyma-project.io/resource":      "deployment",
			"serverless.kyma-project.io/uuid":          "test-uid",
		}, r.Spec.Selector.MatchLabels)
		require.Equal(t, map[string]string{
			"serverless.kyma-project.io/function-name": "test-function-name",
			"serverless.kyma-project.io/managed-by":    "function-controller",
			"serverless.kyma-project.io/resource":      "deployment",
			"serverless.kyma-project.io/uuid":          "test-uid",
			"app.kubernetes.io/name":                   "test-function-name",
			"shtern":                                   "stoic",
			"boyd":                                     "vigilant",
		}, r.Spec.Template.ObjectMeta.Labels)
	})
	t.Run("create annotations based on function and current deployment", func(t *testing.T) {
		d := minimalDeployment()
		d.function.Spec.Annotations = map[string]string{
			"leavitt": "hopeful",
			"pike":    "tender",
		}
		d.function.Status.FunctionAnnotations = map[string]string{
			"dewdney": "intelligent", // this should be removed from deployment
			"leavitt": "hopeful",
		}
		d.clusterDeployment = &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"thompson": "exciting",
							"dewdney":  "zealous", // this should be removed
						},
					},
				},
			},
		}

		r := d.construct()

		require.NotNil(t, r)
		require.Equal(t, map[string]string{
			"proxy.istio.io/config":          "{ \"holdApplicationUntilProxyStarts\": true }",
			"leavitt":                        "hopeful",
			"pike":                           "tender",
			"thompson":                       "exciting",
			"sidecar.istio.io/nativeSidecar": "true",
		}, r.Spec.Template.ObjectMeta.Annotations)
	})
	t.Run("enable native sidecar", func(t *testing.T) {
		d := minimalDeployment()

		r := d.construct()

		require.NotNil(t, r)
		require.Equal(t, map[string]string{
			"proxy.istio.io/config":          "{ \"holdApplicationUntilProxyStarts\": true }",
			"sidecar.istio.io/nativeSidecar": "true",
		}, r.Spec.Template.ObjectMeta.Annotations)
	})
	t.Run("use fixed container name", func(t *testing.T) {
		d := minimalDeployment()

		r := d.construct()

		require.NotNil(t, r)
		require.Equal(t, "function", r.Spec.Template.Spec.Containers[0].Name)
	})
	t.Run("use container image based on function and function configuration", func(t *testing.T) {
		d := &Deployment{
			Deployment: nil,
			functionConfig: &config.FunctionConfig{
				Images: config.ImagesConfig{Python312: "special-test-image"},
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
				`echo "${FUNC_HANDLER_SOURCE}" > handler.py;
PIP_CONFIG_FILE=package-registry-config/pip.conf pip install --user --no-cache-dir -r /kubeless/requirements.txt;
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
	t.Run("doesn't create init container for inline function", func(t *testing.T) {
		d := minimalDeployment()

		r := d.construct()

		require.NotNil(t, r)
		require.Empty(t, r.Spec.Template.Spec.InitContainers)
	})
	t.Run("create init container for git function with data based on function", func(t *testing.T) {
		d := minimalDeployment()
		d.commit = "test-commit"
		d.function.Spec.Source = serverlessv1alpha2.Source{
			GitRepository: &serverlessv1alpha2.GitRepositorySource{
				URL: "wonderful-germain",
				Repository: serverlessv1alpha2.Repository{
					BaseDir:   "recursing-mcnulty",
					Reference: "epic-mendel"}}}

		r := d.construct()

		require.NotNil(t, r)
		require.Len(t, r.Spec.Template.Spec.InitContainers, 1)
		c := r.Spec.Template.Spec.InitContainers[0]
		expectedCommand := []string{"sh", "-c",
			`/app/gitcloner
mkdir /git-repository/src;cp /git-repository/repo/recursing-mcnulty/* /git-repository/src;`}
		require.Equal(t, expectedCommand, c.Command)
	})
}

func TestDeployment_replicas(t *testing.T) {
	t.Run("get replicas from function", func(t *testing.T) {
		d := &Deployment{
			function: &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{
					Replicas: ptr.To[int32](17),
				},
			},
		}

		r := d.replicas()

		assert.Equal(t, int32(17), *r)
	})
	t.Run("get default replicas", func(t *testing.T) {
		d := &Deployment{
			function: &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{},
			},
		}

		r := d.replicas()

		assert.Equal(t, int32(1), *r)
	})
}

func TestDeployment_workingSourcesDir(t *testing.T) {
	tests := []struct {
		name    string
		runtime serverlessv1alpha2.Runtime
		want    string
	}{
		{
			name:    "get working dir for nodejs20",
			runtime: serverlessv1alpha2.NodeJs20,
			want:    "/usr/src/app/function",
		},
		{
			name:    "get working dir for nodejs22",
			runtime: serverlessv1alpha2.NodeJs22,
			want:    "/usr/src/app/function",
		},
		{
			name:    "get working dir for python312",
			runtime: serverlessv1alpha2.Python312,
			want:    "/kubeless",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Deployment{
				function: &serverlessv1alpha2.Function{
					Spec: serverlessv1alpha2.FunctionSpec{
						Runtime: tt.runtime,
					},
				},
			}

			r := d.workingSourcesDir()

			assert.Equal(t, tt.want, r)
		})
	}
}

func TestDeployment_runtimeImage(t *testing.T) {
	c := &config.FunctionConfig{
		Images: config.ImagesConfig{
			NodeJs20:  "image-for-nodejs20",
			NodeJs22:  "image-for-nodejs22",
			Python312: "image-for-python312",
		},
	}
	type fields struct {
		runtime              serverlessv1alpha2.Runtime
		runtimeImageOverride string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "get python312 image from function config",
			fields: fields{
				runtime:              serverlessv1alpha2.Python312,
				runtimeImageOverride: "",
			},
			want: "image-for-python312",
		},
		{
			name: "get nodejs20 image from function config",
			fields: fields{
				runtime:              serverlessv1alpha2.NodeJs20,
				runtimeImageOverride: "",
			},
			want: "image-for-nodejs20",
		},
		{
			name: "get nodejs22 image from function config",
			fields: fields{
				runtime:              serverlessv1alpha2.NodeJs22,
				runtimeImageOverride: "",
			},
			want: "image-for-nodejs22",
		},
		{
			name: "get overridden image name from function",
			fields: fields{
				runtime:              serverlessv1alpha2.NodeJs20,
				runtimeImageOverride: "overridden-image",
			},
			want: "overridden-image",
		},
		{
			name: "get overridden image name from function",
			fields: fields{
				runtime:              serverlessv1alpha2.NodeJs22,
				runtimeImageOverride: "overridden-image",
			},
			want: "overridden-image",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Deployment{
				functionConfig: c,
				function: &serverlessv1alpha2.Function{
					Spec: serverlessv1alpha2.FunctionSpec{
						Runtime:              tt.fields.runtime,
						RuntimeImageOverride: tt.fields.runtimeImageOverride,
					},
				},
			}

			r := d.runtimeImage()

			assert.Equal(t, tt.want, r)
		})
	}
}

func TestDeployment_resourceConfiguration(t *testing.T) {
	rc := &serverlessv1alpha2.ResourceConfiguration{
		Function: &serverlessv1alpha2.ResourceRequirements{
			Resources: &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    k8sresource.MustParse("23m"),
					corev1.ResourceMemory: k8sresource.MustParse("34Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    k8sresource.MustParse("12m"),
					corev1.ResourceMemory: k8sresource.MustParse("24Mi"),
				},
			},
		},
	}
	fc := config.FunctionConfig{
		ResourceConfig: config.ResourceConfig{
			Function: config.FunctionResourceConfig{
				Resources: config.Resources{
					Presets: config.Preset{
						"competent": config.Resource{
							RequestCPU:    config.Quantity{Quantity: k8sresource.MustParse("11m")},
							RequestMemory: config.Quantity{Quantity: k8sresource.MustParse("12Mi")},
							LimitCPU:      config.Quantity{Quantity: k8sresource.MustParse("13m")},
							LimitMemory:   config.Quantity{Quantity: k8sresource.MustParse("14Mi")},
						},
						"quirky": config.Resource{
							RequestCPU:    config.Quantity{Quantity: k8sresource.MustParse("21m")},
							RequestMemory: config.Quantity{Quantity: k8sresource.MustParse("22Mi")},
							LimitCPU:      config.Quantity{Quantity: k8sresource.MustParse("23m")},
							LimitMemory:   config.Quantity{Quantity: k8sresource.MustParse("24Mi")},
						},
						"sad": config.Resource{
							RequestCPU:    config.Quantity{Quantity: k8sresource.MustParse("31m")},
							RequestMemory: config.Quantity{Quantity: k8sresource.MustParse("32Mi")},
							LimitCPU:      config.Quantity{Quantity: k8sresource.MustParse("33m")},
							LimitMemory:   config.Quantity{Quantity: k8sresource.MustParse("34Mi")},
						},
					},
					DefaultPreset: "quirky",
				},
			},
		},
	}
	tests := []struct {
		name           string
		function       *serverlessv1alpha2.Function
		functionConfig config.FunctionConfig
		want           corev1.ResourceRequirements
	}{
		{
			name: "get custom resource configuration from function",
			function: &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: rc,
				},
			},
			want: *rc.Function.Resources,
		},
		{
			name: "get default (empty) resource configuration",
			function: &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{},
			},
			want: corev1.ResourceRequirements{},
		},
		{
			name: "get profile resource configuration from function",
			function: &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{
							Profile: "competent",
						},
					},
				},
			},
			functionConfig: fc,
			want:           fc.ResourceConfig.Function.Resources.Presets["competent"].ToResourceRequirements(),
		},
		{
			name: "get default resource configuration from function config",
			function: &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{},
			},
			functionConfig: fc,
			want:           fc.ResourceConfig.Function.Resources.Presets["quirky"].ToResourceRequirements(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Deployment{
				function:       tt.function,
				functionConfig: &tt.functionConfig,
			}

			r := d.resourceConfiguration()

			require.Equal(t, tt.want, r)
		})
	}
}

func TestDeployment_volumeMounts(t *testing.T) {
	tests := []struct {
		name    string
		runtime serverlessv1alpha2.Runtime
		source  serverlessv1alpha2.Source
		want    []corev1.VolumeMount
	}{
		{
			name:    "build volume mounts for inline nodejs20 based on function",
			runtime: serverlessv1alpha2.NodeJs20,
			source:  serverlessv1alpha2.Source{Inline: &serverlessv1alpha2.InlineSource{Source: "x", Dependencies: "x"}},
			want: []corev1.VolumeMount{
				{
					Name:      "sources",
					MountPath: "/usr/src/app/function",
				},
				{
					Name:      "tmp",
					ReadOnly:  false,
					MountPath: "/tmp",
				},
				{
					Name:      "package-registry-config",
					ReadOnly:  false,
					MountPath: "/usr/src/app/function/package-registry-config/.npmrc",
					SubPath:   ".npmrc",
				},
			},
		},
		{
			name:    "build volume mounts for inline nodejs22 based on function",
			runtime: serverlessv1alpha2.NodeJs22,
			source:  serverlessv1alpha2.Source{Inline: &serverlessv1alpha2.InlineSource{Source: "x", Dependencies: "x"}},
			want: []corev1.VolumeMount{
				{
					Name:      "sources",
					MountPath: "/usr/src/app/function",
				},
				{
					Name:      "tmp",
					ReadOnly:  false,
					MountPath: "/tmp",
				},
				{
					Name:      "package-registry-config",
					ReadOnly:  false,
					MountPath: "/usr/src/app/function/package-registry-config/.npmrc",
					SubPath:   ".npmrc",
				},
			},
		},
		{
			name:    "build volume mounts for inline python312 based on function",
			runtime: serverlessv1alpha2.Python312,
			source:  serverlessv1alpha2.Source{Inline: &serverlessv1alpha2.InlineSource{Source: "x", Dependencies: "x"}},
			want: []corev1.VolumeMount{
				{
					Name:      "sources",
					MountPath: "/kubeless",
				},
				{
					Name:      "tmp",
					ReadOnly:  false,
					MountPath: "/tmp",
				},
				{
					Name:      "local",
					MountPath: "/.local",
				},
				{
					Name:      "package-registry-config",
					ReadOnly:  false,
					MountPath: "/kubeless/package-registry-config/pip.conf",
					SubPath:   "pip.conf",
				},
			},
		},
		{
			name:    "build volume mounts for git nodejs22 based on function",
			runtime: serverlessv1alpha2.NodeJs22,
			source: serverlessv1alpha2.Source{GitRepository: &serverlessv1alpha2.GitRepositorySource{
				URL: "x", Repository: serverlessv1alpha2.Repository{BaseDir: "x", Reference: "x"}}},
			want: []corev1.VolumeMount{
				{
					Name:      "sources",
					MountPath: "/usr/src/app/function",
				},
				{
					Name:      "tmp",
					ReadOnly:  false,
					MountPath: "/tmp",
				},
				{
					Name:      "git-repository",
					ReadOnly:  false,
					MountPath: "/git-repository",
				},
				{
					Name:      "package-registry-config",
					ReadOnly:  false,
					MountPath: "/usr/src/app/function/package-registry-config/.npmrc",
					SubPath:   ".npmrc",
				},
			},
		},
		{
			name:    "build volume mounts for git python312 based on function",
			runtime: serverlessv1alpha2.Python312,
			source: serverlessv1alpha2.Source{GitRepository: &serverlessv1alpha2.GitRepositorySource{
				URL: "x", Repository: serverlessv1alpha2.Repository{BaseDir: "x", Reference: "x"}}},
			want: []corev1.VolumeMount{
				{
					Name:      "sources",
					MountPath: "/kubeless",
				},
				{
					Name:      "tmp",
					ReadOnly:  false,
					MountPath: "/tmp",
				},
				{
					Name:      "git-repository",
					ReadOnly:  false,
					MountPath: "/git-repository",
				},
				{
					Name:      "local",
					MountPath: "/.local",
				},
				{
					Name:      "package-registry-config",
					ReadOnly:  false,
					MountPath: "/kubeless/package-registry-config/pip.conf",
					SubPath:   "pip.conf",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Deployment{
				function: &serverlessv1alpha2.Function{
					Spec: serverlessv1alpha2.FunctionSpec{
						Runtime: tt.runtime,
						Source:  tt.source,
					},
				},
			}

			r := d.volumeMounts()

			assert.Equal(t, tt.want, r)
		})
	}
}

func TestDeployment_volumes(t *testing.T) {
	c := &config.FunctionConfig{
		PackageRegistryConfigSecretName: "test-secret-name",
	}
	tests := []struct {
		name    string
		runtime serverlessv1alpha2.Runtime
		source  serverlessv1alpha2.Source
		want    []corev1.Volume
	}{
		{
			name:    "build volumes for inline nodejs20 based on function",
			runtime: serverlessv1alpha2.NodeJs20,
			source:  serverlessv1alpha2.Source{Inline: &serverlessv1alpha2.InlineSource{Source: "x", Dependencies: "x"}},
			want: []corev1.Volume{
				{
					Name: "sources",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "package-registry-config",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "test-secret-name",
							Optional:   ptr.To[bool](true),
						},
					},
				},
				{
					Name: "tmp",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
		{
			name:    "build volumes for inline nodejs22 based on function",
			runtime: serverlessv1alpha2.NodeJs22,
			source:  serverlessv1alpha2.Source{Inline: &serverlessv1alpha2.InlineSource{Source: "x", Dependencies: "x"}},
			want: []corev1.Volume{
				{
					Name: "sources",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "package-registry-config",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "test-secret-name",
							Optional:   ptr.To[bool](true),
						},
					},
				},
				{
					Name: "tmp",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
		{
			name:    "build volumes for inline python312 based on function",
			runtime: serverlessv1alpha2.Python312,
			source:  serverlessv1alpha2.Source{Inline: &serverlessv1alpha2.InlineSource{Source: "x", Dependencies: "x"}},
			want: []corev1.Volume{
				{
					Name: "sources",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "package-registry-config",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "test-secret-name",
							Optional:   ptr.To[bool](true),
						},
					},
				},
				{
					Name: "tmp",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "local",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
		{
			name:    "build volumes for git nodejs22 based on function",
			runtime: serverlessv1alpha2.NodeJs22,
			source: serverlessv1alpha2.Source{GitRepository: &serverlessv1alpha2.GitRepositorySource{
				URL: "x", Repository: serverlessv1alpha2.Repository{BaseDir: "x", Reference: "x"}}},
			want: []corev1.Volume{
				{
					Name: "sources",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "package-registry-config",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "test-secret-name",
							Optional:   ptr.To[bool](true),
						},
					},
				},
				{
					Name: "tmp",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "git-repository",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
		{
			name:    "build volumes for inline python312 based on function",
			runtime: serverlessv1alpha2.Python312,
			source: serverlessv1alpha2.Source{GitRepository: &serverlessv1alpha2.GitRepositorySource{
				URL: "x", Repository: serverlessv1alpha2.Repository{BaseDir: "x", Reference: "x"}}},
			want: []corev1.Volume{
				{
					Name: "sources",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "package-registry-config",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "test-secret-name",
							Optional:   ptr.To[bool](true),
						},
					},
				},
				{
					Name: "tmp",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "git-repository",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "local",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Deployment{
				functionConfig: c,
				function: &serverlessv1alpha2.Function{
					Spec: serverlessv1alpha2.FunctionSpec{
						Runtime: tt.runtime,
						Source:  tt.source,
					},
				},
			}

			r := d.volumes()

			assert.Equal(t, tt.want, r)
		})
	}
}

func TestDeployment_deploymentSecretVolumes(t *testing.T) {
	tests := []struct {
		name             string
		secretMounts     []serverlessv1alpha2.SecretMount
		wantVolumes      []corev1.Volume
		wantVolumeMounts []corev1.VolumeMount
	}{
		{
			name: "build secret volumes based on function",
			secretMounts: []serverlessv1alpha2.SecretMount{
				{
					SecretName: "secret-name-1",
					MountPath:  "mount-path-1",
				},
				{
					SecretName: "secret-name-2",
					MountPath:  "mount-path-2",
				},
			},
			wantVolumes: []corev1.Volume{
				{
					Name: "secret-name-1",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  "secret-name-1",
							DefaultMode: ptr.To[int32](0666),
							Optional:    ptr.To[bool](false),
						},
					},
				},
				{
					Name: "secret-name-2",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  "secret-name-2",
							DefaultMode: ptr.To[int32](0666),
							Optional:    ptr.To[bool](false),
						},
					},
				},
			},
			wantVolumeMounts: []corev1.VolumeMount{
				{
					Name:      "secret-name-1",
					ReadOnly:  true,
					MountPath: "mount-path-1",
				},
				{
					Name:      "secret-name-2",
					ReadOnly:  true,
					MountPath: "mount-path-2",
				},
			},
		},
		{
			name:             "build empty secret volumes based on function",
			secretMounts:     []serverlessv1alpha2.SecretMount{},
			wantVolumes:      []corev1.Volume{},
			wantVolumeMounts: []corev1.VolumeMount{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Deployment{
				function: &serverlessv1alpha2.Function{
					Spec: serverlessv1alpha2.FunctionSpec{
						SecretMounts: tt.secretMounts,
					},
				},
			}
			rV, rVM := d.deploymentSecretVolumes()
			assert.Equal(t, tt.wantVolumes, rV)
			assert.Equal(t, tt.wantVolumeMounts, rVM)
		})
	}
}

func TestDeployment_envs(t *testing.T) {
	tests := []struct {
		name     string
		function *serverlessv1alpha2.Function
		fnConfig config.FunctionConfig
		want     []corev1.EnvVar
	}{
		{
			name: "build envs based on inline nodejs20 function",
			function: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "function-namespace",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs20,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "function-source",
							Dependencies: "function-dependencies",
						},
					},
				},
			},
			want: []corev1.EnvVar{
				{
					Name:  "SERVICE_NAMESPACE",
					Value: "function-namespace",
				},
				{
					Name:  "FUNC_HANDLER_SOURCE",
					Value: "function-source",
				},
				{
					Name:  "FUNC_HANDLER_DEPENDENCIES",
					Value: "function-dependencies",
				},
				{
					Name:  "TRACE_COLLECTOR_ENDPOINT",
					Value: "test-trace-collector-endpoint",
				},
				{
					Name:  "PUBLISHER_PROXY_ADDRESS",
					Value: "test-proxy-address",
				},
			},
		},
		{
			name: "build envs based on inline nodejs22 function",
			function: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "function-namespace",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs22,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "function-source",
							Dependencies: "function-dependencies",
						},
					},
				},
			},
			want: []corev1.EnvVar{
				{
					Name:  "SERVICE_NAMESPACE",
					Value: "function-namespace",
				},
				{
					Name:  "FUNC_HANDLER_SOURCE",
					Value: "function-source",
				},
				{
					Name:  "FUNC_HANDLER_DEPENDENCIES",
					Value: "function-dependencies",
				},
				{
					Name:  "TRACE_COLLECTOR_ENDPOINT",
					Value: "test-trace-collector-endpoint",
				},
				{
					Name:  "PUBLISHER_PROXY_ADDRESS",
					Value: "test-proxy-address",
				},
			},
		},
		{
			name: "build envs based on git nodejs22 function",
			function: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "function-namespace",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs22,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: "/some/url",
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "/some/dir",
								Reference: "some-reference",
							},
						},
					},
				},
			},
			want: []corev1.EnvVar{
				{
					Name:  "SERVICE_NAMESPACE",
					Value: "function-namespace",
				},
				{
					Name:  "TRACE_COLLECTOR_ENDPOINT",
					Value: "test-trace-collector-endpoint",
				},
				{
					Name:  "PUBLISHER_PROXY_ADDRESS",
					Value: "test-proxy-address",
				},
			},
		},
		{
			name: "build envs based on inline python312 function",
			function: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "function-namespace",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "function-source-py",
							Dependencies: "function-dependencies-py",
						},
					},
				},
			},
			want: []corev1.EnvVar{
				{
					Name:  "SERVICE_NAMESPACE",
					Value: "function-namespace",
				},
				{
					Name:  "FUNC_HANDLER_SOURCE",
					Value: "function-source-py",
				},
				{
					Name:  "FUNC_HANDLER_DEPENDENCIES",
					Value: "function-dependencies-py",
				},
				{
					Name:  "TRACE_COLLECTOR_ENDPOINT",
					Value: "test-trace-collector-endpoint",
				},
				{
					Name:  "PUBLISHER_PROXY_ADDRESS",
					Value: "test-proxy-address",
				},
				{
					Name:  "MOD_NAME",
					Value: "handler",
				},
				{
					Name:  "FUNC_HANDLER",
					Value: "main",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Deployment{
				function: tt.function,
				functionConfig: &config.FunctionConfig{
					FunctionPublisherProxyAddress:  "test-proxy-address",
					FunctionTraceCollectorEndpoint: "test-trace-collector-endpoint",
				},
			}

			r := d.envs()

			assert.ElementsMatch(t, tt.want, r)
		})
	}
}

func TestDeployment_runtimeCommand(t *testing.T) {
	tests := []struct {
		name     string
		function *serverlessv1alpha2.Function
		want     string
	}{
		{
			name: "build runtime command for inline python312 without dependencies",
			function: &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source: "function-source",
						},
					},
				},
			},
			want: `echo "${FUNC_HANDLER_SOURCE}" > handler.py;
PIP_CONFIG_FILE=package-registry-config/pip.conf pip install --user --no-cache-dir -r /kubeless/requirements.txt;
cd ..;
python /kubeless.py;`,
		},
		{
			name: "build runtime command for inline python312 with dependencies",
			function: &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "function-source",
							Dependencies: "function-dependencies",
						},
					},
				},
			},
			want: `echo "${FUNC_HANDLER_SOURCE}" > handler.py;
echo "${FUNC_HANDLER_DEPENDENCIES}" > requirements.txt;
PIP_CONFIG_FILE=package-registry-config/pip.conf pip install --user --no-cache-dir -r /kubeless/requirements.txt;
cd ..;
python /kubeless.py;`,
		},
		{
			name: "build runtime command for git python312",
			function: &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: "/some/url",
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "/some/dir",
								Reference: "some-reference",
							},
						},
					},
				},
			},
			want: `cp /git-repository/src/* .;
PIP_CONFIG_FILE=package-registry-config/pip.conf pip install --user --no-cache-dir -r /kubeless/requirements.txt;
cd ..;
python /kubeless.py;`,
		},
		{
			name: "build runtime command for inline nodejs20 without dependencies",
			function: &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs20,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source: "function-source",
						},
					},
				},
			},
			want: `echo "{}" > package.json;
echo "${FUNC_HANDLER_SOURCE}" > handler.js;
npm install --prefer-offline --no-audit --progress=false;
cd ..;
npm start;`,
		},
		{
			name: "build runtime command for inline nodejs20 with dependencies",
			function: &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs20,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "function-source",
							Dependencies: "function-dependencies",
						},
					},
				},
			},
			want: `echo "{}" > package.json;
echo "${FUNC_HANDLER_SOURCE}" > handler.js;
echo "${FUNC_HANDLER_DEPENDENCIES}" > package.json;
npm install --prefer-offline --no-audit --progress=false;
cd ..;
npm start;`,
		},
		{
			name: "build runtime command for git nodejs20",
			function: &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs20,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: "/some/url",
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "/some/dir",
								Reference: "some-reference",
							},
						},
					},
				},
			},
			want: `echo "{}" > package.json;
cp /git-repository/src/* .;
npm install --prefer-offline --no-audit --progress=false;
cd ..;
npm start;`,
		},
		{
			name: "build runtime command for inline nodejs22 without dependencies",
			function: &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs22,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source: "function-source",
						},
					},
				},
			},
			want: `echo "{}" > package.json;
echo "${FUNC_HANDLER_SOURCE}" > handler.js;
npm install --prefer-offline --no-audit --progress=false;
cd ..;
npm start;`,
		},
		{
			name: "build runtime command for inline nodejs22 with dependencies",
			function: &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs22,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "function-source",
							Dependencies: "function-dependencies",
						},
					},
				},
			},
			want: `echo "{}" > package.json;
echo "${FUNC_HANDLER_SOURCE}" > handler.js;
echo "${FUNC_HANDLER_DEPENDENCIES}" > package.json;
npm install --prefer-offline --no-audit --progress=false;
cd ..;
npm start;`,
		},
		{
			name: "build runtime command for git nodejs22",
			function: &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs22,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: "/some/url",
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "/some/dir",
								Reference: "some-reference",
							},
						},
					},
				},
			},
			want: `echo "{}" > package.json;
cp /git-repository/src/* .;
npm install --prefer-offline --no-audit --progress=false;
cd ..;
npm start;`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Deployment{
				function: tt.function,
			}

			r := d.runtimeCommand()

			assert.Equal(t, tt.want, r)
		})
	}
}

func minimalFunction() *serverlessv1alpha2.Function {
	return &serverlessv1alpha2.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-function-name",
			Namespace: "test-function-namespace",
			UID:       "test-uid",
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
		Images: config.ImagesConfig{
			Python312: "test-image-python312",
		},
	}
}

func minimalDeployment() *Deployment {
	return &Deployment{
		Deployment:     nil,
		functionConfig: minimalFunctionConfig(),
		function:       minimalFunction(),
	}
}
