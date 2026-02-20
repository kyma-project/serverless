package validator

import (
	"fmt"
	"testing"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/config"
	"github.com/kyma-project/serverless/components/common/fips"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mockFipsChecker(enabled bool) fips.FipsChecker {
	return func() bool {
		return enabled
	}
}

func TestNewFunctionValidator(t *testing.T) {
	t.Run("create function validator", func(t *testing.T) {
		f := &serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "compassionate-villani-name",
				Namespace: "vigorous-jang-ns"}}

		r := New(f, config.FunctionConfig{}, mockFipsChecker(false))

		require.NotNil(t, r)
		require.NotNil(t, r.instance)
		require.Equal(t, "compassionate-villani-name", r.instance.GetName())
		require.Equal(t, "vigorous-jang-ns", r.instance.GetNamespace())
	})
}

func Test_functionValidator_Validate(t *testing.T) {
	t.Run("when function is valid should return empty list", func(t *testing.T) {
		v := New(&serverlessv1alpha2.Function{}, config.FunctionConfig{}, mockFipsChecker(false))

		r := v.Validate()

		require.Len(t, r, 0)
	})
	t.Run("when function is invalid should return list with all errors", func(t *testing.T) {
		f := &serverlessv1alpha2.Function{
			Spec: serverlessv1alpha2.FunctionSpec{
				Env:     []corev1.EnvVar{{Name: "goofy-kare;;;;;"}},
				Source:  serverlessv1alpha2.Source{Inline: &serverlessv1alpha2.InlineSource{}},
				Runtime: "upbeat-boyd",
			}}

		v := New(f, config.FunctionConfig{}, mockFipsChecker(false))

		r := v.Validate()

		require.ElementsMatch(t, []string{
			"spec.env: goofy-kare;;;;;. Err: a valid environment variable name must consist of alphabetic characters, digits, '_', '-', or '.', and must not start with a digit (e.g. 'my.env-name',  or 'MY_ENV.NAME',  or 'MyEnvName1', regex used for validation is '[-._a-zA-Z][-._a-zA-Z0-9]*')",
			"invalid source.inline.dependencies value: cannot find runtime: upbeat-boyd",
			"invalid runtime value: cannot find runtime: upbeat-boyd",
		}, r)
	})
}

func Test_functionValidator_validateEnvs(t *testing.T) {
	tests := []struct {
		name string
		envs []corev1.EnvVar
		want []string
	}{
		{
			name: "when empty envs then no errors",
			envs: []corev1.EnvVar{},
			want: []string{},
		},
		{
			name: "when valid env names then no errors",
			envs: []corev1.EnvVar{
				{Name: "sad-bose", Value: "!@#$%^&*()"},
				{Name: "wonderful-hypatia", Value: "`~-=_+[]{};':,.<>"},
			},
			want: []string{},
		},
		{
			name: "when env name is invalid then produces error for it",
			envs: []corev1.EnvVar{
				{Name: "objective-gauss"},
				{Name: "2lucid-volhard"},
				{Name: "laughing-keldysh"},
			},
			want: []string{
				"spec.env: 2lucid-volhard. Err: a valid environment variable name must consist of alphabetic characters, digits, '_', '-', or '.', and must not start with a digit (e.g. 'my.env-name',  or 'MY_ENV.NAME',  or 'MyEnvName1', regex used for validation is '[-._a-zA-Z][-._a-zA-Z0-9]*')",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{
					Env: tt.envs,
				}}

			v := New(f, config.FunctionConfig{}, mockFipsChecker(false))
			r := v.validateEnvs()
			require.ElementsMatch(t, tt.want, r)
		})
	}
}

func Test_functionValidator_validateInlineDeps(t *testing.T) {
	tests := []struct {
		name string
		spec serverlessv1alpha2.FunctionSpec
		want []string
	}{
		{
			name: "when empty inline source then no errors",
			spec: serverlessv1alpha2.FunctionSpec{},
			want: []string{},
		},
		{
			name: "when unknown runtime then return error",
			spec: serverlessv1alpha2.FunctionSpec{
				Runtime: "pedantic-lewin",
				Source: serverlessv1alpha2.Source{
					Inline: &serverlessv1alpha2.InlineSource{},
				},
			},
			want: []string{
				"invalid source.inline.dependencies value: cannot find runtime: pedantic-lewin",
			},
		},
		{
			name: "when python runtime then no errors",
			spec: serverlessv1alpha2.FunctionSpec{
				Runtime: serverlessv1alpha2.Python312,
				Source: serverlessv1alpha2.Source{
					Inline: &serverlessv1alpha2.InlineSource{
						Source:       "sweet-goldstine",
						Dependencies: "`1234567890-=qwertyuiop[]asdfghjkl;'\\zxcvbnm,./~!@#$%^&*()_+{}:|<>?",
					},
				},
			},
			want: []string{},
		},
		{
			name: "when js runtime with invalid dependencies then return error",
			spec: serverlessv1alpha2.FunctionSpec{
				Runtime: serverlessv1alpha2.NodeJs24,
				Source: serverlessv1alpha2.Source{
					Inline: &serverlessv1alpha2.InlineSource{
						Source:       "intelligent-fermi",
						Dependencies: "`1234567890-=qwertyuiop[]asdfghjkl;'\\zxcvbnm,./~!@#$%^&*()_+{}:|<>?",
					},
				},
			},
			want: []string{
				"invalid source.inline.dependencies value: deps should start with '{' and end with '}'",
			},
		},
		{
			name: "when js runtime with valid dependencies then no errors",
			spec: serverlessv1alpha2.FunctionSpec{
				Runtime: serverlessv1alpha2.NodeJs24,
				Source: serverlessv1alpha2.Source{
					Inline: &serverlessv1alpha2.InlineSource{
						Source:       "epic-swirles",
						Dependencies: "{epic-swirles}",
					},
				},
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &serverlessv1alpha2.Function{
				Spec: tt.spec,
			}

			v := New(f, config.FunctionConfig{}, mockFipsChecker(false))
			r := v.validateInlineDeps()
			require.ElementsMatch(t, tt.want, r)
		})
	}
}

func Test_functionValidator_validateRuntime(t *testing.T) {
	type testData struct {
		name    string
		runtime serverlessv1alpha2.Runtime
		want    []string
	}
	tests := []testData{
		{
			name:    "when empty runtime then no errors",
			runtime: "",
			want:    []string{},
		},
		{
			name:    "when unknown runtime then return error",
			runtime: "practical-panini",
			want: []string{
				"invalid runtime value: cannot find runtime: practical-panini",
			},
		},
	}
	for _, runtime := range []serverlessv1alpha2.Runtime{serverlessv1alpha2.NodeJs20, serverlessv1alpha2.NodeJs22, serverlessv1alpha2.NodeJs24, serverlessv1alpha2.Python312} {
		tests = append(tests, testData{
			name:    fmt.Sprintf("when %s then no errors", runtime),
			runtime: runtime,
			want:    []string{},
		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &serverlessv1alpha2.Function{
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: tt.runtime,
				},
			}

			v := New(f, config.FunctionConfig{}, mockFipsChecker(false))
			r := v.validateRuntime()
			require.ElementsMatch(t, tt.want, r)
		})
	}
}

func Test_validator_validateSecretMounts(t *testing.T) {
	type testData struct {
		name         string
		secretMounts []serverlessv1alpha2.SecretMount
		want         []string
	}
	tests := []testData{
		{
			name:         "when no secret mounts then no errors",
			secretMounts: []serverlessv1alpha2.SecretMount{},
			want:         []string{},
		},
		{
			name: "when secret name is invalid then return error",
			secretMounts: []serverlessv1alpha2.SecretMount{
				{SecretName: "invalid_secret_name@#!"},
			},
			want: []string{
				"invalid spec.secretMounts: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
			},
		},
		{
			name: "when secret names are not unique then return error",
			secretMounts: []serverlessv1alpha2.SecretMount{
				{SecretName: "valid-secret"},
				{SecretName: "valid-secret"},
			},
			want: []string{
				"invalid spec.secretMounts: [secretNames should be unique]",
			},
		},
		{
			name: "when secret names are valid and unique then no errors",
			secretMounts: []serverlessv1alpha2.SecretMount{
				{SecretName: "valid-secret-chlebek"},
				{SecretName: "valid-secret-chleb"},
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &validator{
				instance: &serverlessv1alpha2.Function{
					Spec: serverlessv1alpha2.FunctionSpec{
						SecretMounts: tt.secretMounts,
					},
				},
			}
			got := v.validateSecretMounts()
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func Test_validator_validateFunctionLabels(t *testing.T) {
	type testData struct {
		name   string
		labels map[string]string
		want   []string
	}
	tests := []testData{
		{
			name:   "when no labels then no errors",
			labels: map[string]string{},
			want:   []string{},
		},
		{
			name: "when valid labels then no errors",
			labels: map[string]string{
				"valid-label-1": "chlebek",
				"valid-label-2": "chlebek2",
			},
			want: []string{},
		},
		{
			name: "when invalid label key then return error",
			labels: map[string]string{
				"Invalid_Label@Key!": "value",
			},
			want: []string{
				"spec.labels: Invalid value: \"Invalid_Label@Key!\": name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')",
			},
		},
		{
			name: "when invalid label value then return error",
			labels: map[string]string{
				"valid-label": "Invalid_ChlEbek!",
			},
			want: []string{
				"spec.labels: Invalid value: \"Invalid_ChlEbek!\": a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &validator{
				instance: &serverlessv1alpha2.Function{
					Spec: serverlessv1alpha2.FunctionSpec{
						Labels: tt.labels,
					},
				},
			}
			got := v.validateFunctionLabels()
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func Test_validator_validateFunctionAnnotations(t *testing.T) {
	type testData struct {
		name        string
		annotations map[string]string
		want        []string
	}
	tests := []testData{
		{
			name:        "when no annotations then no errors",
			annotations: map[string]string{},
			want:        []string{},
		},
		{
			name: "when valid annotations then no errors",
			annotations: map[string]string{
				"valid-annotation-1": "chlebek",
				"valid-annotation-2": "chleb",
			},
			want: []string{},
		},
		{
			name: "when invalid annotation key then return error",
			annotations: map[string]string{
				"Invalid_Annotation@Key!": "value",
			},
			want: []string{
				"spec.annotations: Invalid value: \"Invalid_Annotation@Key!\": name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')",
			},
		},
		{
			name: "when invalid annotation value then no errors",
			annotations: map[string]string{
				"valid-annotation": "Invalid_Value!@",
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &validator{
				instance: &serverlessv1alpha2.Function{
					Spec: serverlessv1alpha2.FunctionSpec{
						Annotations: tt.annotations,
					},
				},
			}
			got := v.validateFunctionAnnotations()
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func Test_validator_validateGitRepoURL(t *testing.T) {
	type testData struct {
		name string
		URL  string
		want []string
	}
	tests := []testData{
		{
			name: "when Git repo URL is valid SSH then no errors",
			URL:  "git@github.com:user/repo.git",
			want: []string{},
		},
		{
			name: "when Git repo URL is valid HTTPS then no errors",
			URL:  "https://github.com/user/repo.git",
			want: []string{},
		},
		{
			name: "when Git repo URL is invalid then return error",
			URL:  "invalid-url",
			want: []string{
				"source.gitRepository.URL: parse \"invalid-url\": invalid URI for request",
			},
		},
		{
			name: "when Git repo URL is empty then return error",
			URL:  "source.gitRepository.URL: parse \\\"\\\": empty url",
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &validator{
				instance: &serverlessv1alpha2.Function{
					Spec: serverlessv1alpha2.FunctionSpec{
						Source: serverlessv1alpha2.Source{
							GitRepository: &serverlessv1alpha2.GitRepositorySource{
								URL: tt.URL,
							},
						},
					},
				},
			}
			got := v.validateGitRepoURL()
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func Test_validator_validateFips(t *testing.T) {
	type testData struct {
		name     string
		fipsMode bool
		URL      string
		want     []string
	}
	tests := []testData{
		{
			name:     "FIPS enabled with SSH URL should return error",
			fipsMode: true,
			URL:      "git@github.com:user/repo.git",
			want:     []string{"SSH source.gitRepository.URL is not allowed in FIPS mode"},
		},
		{
			name:     "FIPS enabled with HTTPS URL should return no errors",
			fipsMode: true,
			URL:      "https://github.com/user/repo.git",
			want:     []string{},
		},
		{
			name:     "FIPS enabled with HTTP URL should return no errors",
			fipsMode: true,
			URL:      "http://github.com/user/repo.git",
			want:     []string{},
		},
		{
			name:     "FIPS disabled with SSH URL should return no errors",
			fipsMode: false,
			URL:      "git@github.com:user/repo.git",
			want:     []string{},
		},
		{
			name:     "FIPS disabled with HTTPS URL should return no errors",
			fipsMode: false,
			URL:      "https://github.com/user/repo.git",
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &validator{
				instance: &serverlessv1alpha2.Function{
					Spec: serverlessv1alpha2.FunctionSpec{
						Source: serverlessv1alpha2.Source{
							GitRepository: &serverlessv1alpha2.GitRepositorySource{
								URL: tt.URL,
							},
						},
					},
				},
				checkFips: mockFipsChecker(tt.fipsMode),
			}
			got := v.validateFips()
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func Test_validator_validateFunctionResources(t *testing.T) {
	type testData struct {
		name       string
		resources  *corev1.ResourceRequirements
		minCPU     resource.Quantity
		minMemory  resource.Quantity
		wantErrors []string
	}
	tests := []testData{
		{
			name:       "when resources are nil then no errors",
			resources:  nil,
			minCPU:     resource.MustParse("10m"),
			minMemory:  resource.MustParse("16Mi"),
			wantErrors: []string{},
		},
		{
			name: "when resources meet minimum requirements then no errors",
			resources: &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("64Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("32Mi"),
				},
			},
			minCPU:     resource.MustParse("10m"),
			minMemory:  resource.MustParse("16Mi"),
			wantErrors: []string{},
		},
		{
			name: "when CPU limit is below minimum then return error",
			resources: &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("5m"),
					corev1.ResourceMemory: resource.MustParse("64Mi"),
				},
			},
			minCPU:    resource.MustParse("10m"),
			minMemory: resource.MustParse("16Mi"),
			wantErrors: []string{
				"Function limits cpu(5m) should be higher than minimal value (10m)",
			},
		},
		{
			name: "when memory request is below minimum then return error",
			resources: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("8Mi"),
				},
			},
			minCPU:    resource.MustParse("10m"),
			minMemory: resource.MustParse("16Mi"),
			wantErrors: []string{
				"Function request memory(8Mi) should be higher than minimal value (16Mi)",
			},
		},
		{
			name: "when requests exceed limits then return errors",
			resources: &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("64Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
			minCPU:    resource.MustParse("10m"),
			minMemory: resource.MustParse("16Mi"),
			wantErrors: []string{
				"Function limits cpu(100m) should be higher than requests cpu(200m)",
				"Function limits memory(64Mi) should be higher than requests memory(128Mi)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &validator{
				instance: &serverlessv1alpha2.Function{
					Spec: serverlessv1alpha2.FunctionSpec{
						ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
							Function: &serverlessv1alpha2.ResourceRequirements{
								Resources: tt.resources,
							},
						},
					},
				},
				fnConfig: config.FunctionConfig{
					ResourceConfig: config.ResourceConfig{
						Function: config.FunctionResourceConfig{
							Resources: config.Resources{
								MinRequestCPU:    config.Quantity{Quantity: tt.minCPU},
								MinRequestMemory: config.Quantity{Quantity: tt.minMemory},
							},
						},
					},
				},
			}
			got := v.validateFunctionResources()
			require.ElementsMatch(t, tt.wantErrors, got)
		})
	}
}
