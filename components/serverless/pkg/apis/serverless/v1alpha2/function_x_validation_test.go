package v1alpha2_test

import (
	"context"
	"github.com/kyma-project/kyma/components/function-controller/internal/testenv"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"testing"
)

func Test_XKubernetesValidations_Valid(t *testing.T) {
	fixMetadata := metav1.ObjectMeta{
		GenerateName: "test",
		Namespace:    "test",
	}
	ctx := context.TODO()
	k8sClient, testEnv := testenv.Start(t)
	defer testenv.Stop(t, testEnv)

	testNs := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
	}
	err := k8sClient.Create(ctx, &testNs)
	require.NoError(t, err)
	//GIVEN
	testCases := map[string]struct {
		fn *serverlessv1alpha2.Function
	}{
		"Profile set only for function": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{Function: &serverlessv1alpha2.ResourceRequirements{
						Profile: "XS",
					}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"Profile set only for buildJob": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{
							Profile: "fast",
						}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"Build profile local-dev": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{
							Profile: "local-dev",
						}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"Build profile slow": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{
							Profile: "slow",
						}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"Build profile normal": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{
							Profile: "normal",
						}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"Build profile fast": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{
							Profile: "fast",
						}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"Function profile S": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{
							Profile: "S",
						}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"Function profile M": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{
							Profile: "M",
						}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"Function profile L": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{
							Profile: "L",
						}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"Function profile XL": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{
							Profile: "XL",
						}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"Resource set only for buildJob": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{}}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"Resource set only for function": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{}}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"labels has value with special characters similar to restricted one ": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Labels: map[string]string{
						"serverless$kyma-project#io/abc": "labelValue",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"labels has value restricted domain without path ": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Labels: map[string]string{
						"serverless.kyma-project.io": "labelValue",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"labels not use restricted domain": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Labels: map[string]string{
						"my.label.com": "labelValue",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"similar label not use restricted domain": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Labels: map[string]string{
						"serverless.kyma-project.label.com": "labelValue",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"annotations has value with special characters similar to restricted one ": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Annotations: map[string]string{
						"serverless$kyma-project#io/abc": "labelValue",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"annotations has value restricted domain without path ": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Annotations: map[string]string{
						"serverless.kyma-project.io": "labelValue",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"annotations not use restricted domain": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Annotations: map[string]string{
						"my.label.com": "labelValue",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"annotations label not use restricted domain": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Annotations: map[string]string{
						"serverless.kyma-project.label.com": "labelValue",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
		},
		"allowed runtime: nodejs16": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs16,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
				},
			},
		},
		"allowed runtime: nodejs18": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs18,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
				},
			},
		},
		"allowed runtime: python39": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python39,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
				},
			},
		},
		"allowed envs": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python39,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Env: []corev1.EnvVar{{Name: "TEST_ENV"}, {Name: "MY_ENV"}},
				},
			},
		},
		"gitRepository used as source": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python39,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{}},
				},
			},
		},
		"label value lenght is 63": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python39,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{}},
					Labels: map[string]string{
						strings.Repeat("a", 63): "test",
					},
				},
			},
		},

		"secretMount": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python39,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{}},
					SecretMounts: []serverlessv1alpha2.SecretMount{{MountPath: "/path", SecretName: "secret-name"}},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			//WHEN
			err := k8sClient.Create(ctx, tc.fn)
			//THEN
			require.NoError(t, err)
		})
	}
}
func Test_XKubernetesValidations_Invalid(t *testing.T) {
	fixMetadata := metav1.ObjectMeta{
		GenerateName: "test",
		Namespace:    "test",
	}
	ctx := context.TODO()
	k8sClient, testEnv := testenv.Start(t)
	defer testenv.Stop(t, testEnv)
	testNs := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
	}
	err := k8sClient.Create(ctx, &testNs)
	require.NoError(t, err)

	//GIVEN
	testCases := map[string]struct {
		fn             *serverlessv1alpha2.Function
		expectedErrMsg string
		fieldPath      string
		expectedCause  metav1.CauseType
	}{
		"Resource and Profiles used together in function": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{Function: &serverlessv1alpha2.ResourceRequirements{
						Profile:   "L",
						Resources: &corev1.ResourceRequirements{},
					}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			expectedErrMsg: "Use profile or resources",
			fieldPath:      "spec.resourceConfiguration.function",
		},
		"Resource and Profiles used together in buildJob": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{Build: &serverlessv1alpha2.ResourceRequirements{
						Profile:   "fast",
						Resources: &corev1.ResourceRequirements{},
					}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			expectedErrMsg: "Use profile or resources",
			fieldPath:      "spec.resourceConfiguration.build",
		},
		"Invalid profile in build": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{
							Profile: "LL",
						}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			expectedErrMsg: "Invalid profile, please use one of: [",
			fieldPath:      "spec.resourceConfiguration.build",
		},
		"Invalid profile in function": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{
							Profile: "SS",
						}},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			expectedErrMsg: "Invalid profile, please use one of: [",
			fieldPath:      "spec.resourceConfiguration.function",
		},
		"labels use exact restricted domain": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Labels: map[string]string{
						"serverless.kyma-project.io/": "labelValue",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			fieldPath:      "spec.labels",
			expectedErrMsg: "Labels has key starting with ",
		},
		"labels use restricted domain": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Labels: map[string]string{
						"serverless.kyma-project.io/abc": "labelValue",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			fieldPath:      "spec.labels",
			expectedErrMsg: "Labels has key starting with ",
		},
		"labels has many values with incorrect one": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Labels: map[string]string{
						"app":                            "mySuperApp",
						"serverless.kyma-project.io/abc": "labelValue",
						"service":                        "mySvc",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			fieldPath:      "spec.labels",
			expectedErrMsg: "Labels has key starting with ",
		},
		"annotations use exact restricted domain": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Annotations: map[string]string{
						"serverless.kyma-project.io/": "labelValue",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			fieldPath:      "spec.annotations",
			expectedErrMsg: "Annotations has key starting with ",
		},
		"annotations has many values with incorrect one": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Annotations: map[string]string{
						"app":                            "mySuperApp",
						"serverless.kyma-project.io/abc": "labelValue",
						"service":                        "mySvc",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			fieldPath:      "spec.annotations",
			expectedErrMsg: "Annotations has key starting with ",
		},
		"annotations use restricted domain": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Annotations: map[string]string{
						"serverless.kyma-project.io/abc": "labelValue",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			fieldPath:      "spec.annotations",
			expectedErrMsg: "Annotations has key starting with ",
		},
		"disallowed runtime: custom": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Runtime("custom"),
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueNotSupported,
			fieldPath:      "spec.runtime",
			expectedErrMsg: `Unsupported value: "custom"`,
		},
		"reserved env: FUNC_RUNTIME": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
					Env: []corev1.EnvVar{
						{Name: "TEST2"},
						{Name: "FUNC_RUNTIME"},
						{Name: "TEST"},
					},
				},
			},
			expectedErrMsg: "Following envs are reserved",
			fieldPath:      "spec.env",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"reserved env: FUNC_HANDLER": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
					Env: []corev1.EnvVar{
						{Name: "TEST2"},
						{Name: "FUNC_HANDLER"},
						{Name: "TEST"},
					},
				},
			},
			expectedErrMsg: "Following envs are reserved",
			fieldPath:      "spec.env",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"reserved env: FUNC_PORT": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
					Env: []corev1.EnvVar{
						{Name: "TEST2"},
						{Name: "FUNC_PORT"},
						{Name: "TEST"},
					},
				},
			},
			expectedErrMsg: "Following envs are reserved",
			fieldPath:      "spec.env",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"reserved env: MOD_NAME": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
					Env: []corev1.EnvVar{
						{Name: "TEST2"},
						{Name: "MOD_NAME"},
						{Name: "TEST"},
					},
				},
			},
			expectedErrMsg: "Following envs are reserved",
			fieldPath:      "spec.env",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"reserved env: NODE_PATH": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
					Env: []corev1.EnvVar{
						{Name: "TEST2"},
						{Name: "NODE_PATH"},
						{Name: "TEST"},
					},
				},
			},
			expectedErrMsg: "Following envs are reserved",
			fieldPath:      "spec.env",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"reserved env: PYTHONPATH": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{}},
					Runtime: serverlessv1alpha2.Python39,
					Env: []corev1.EnvVar{
						{Name: "TEST2"},
						{Name: "PYTHONPATH"},
						{Name: "TEST"},
					},
				},
			},
			expectedErrMsg: "Following envs are reserved",
			fieldPath:      "spec.env",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"GitRepository and Inline source used together": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python39,
					Source: serverlessv1alpha2.Source{
						GitRepository: nil,
						Inline:        nil,
					},
				},
			},
			expectedErrMsg: "Use GitRepository or Inline source",
			fieldPath:      "spec.source",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"Secret Mount name is empty": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python39,
					Source:  serverlessv1alpha2.Source{Inline: &serverlessv1alpha2.InlineSource{}},
					SecretMounts: []serverlessv1alpha2.SecretMount{
						{SecretName: "", MountPath: "/path"},
					},
				},
			},
			expectedErrMsg: "should be at least 1 chars long",
			fieldPath:      "spec.secretMounts[0].secretName",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"Secret Mount path is empty": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python39,
					Source:  serverlessv1alpha2.Source{Inline: &serverlessv1alpha2.InlineSource{}},
					SecretMounts: []serverlessv1alpha2.SecretMount{
						{SecretName: "my-secret", MountPath: ""},
					},
				},
			},
			expectedErrMsg: "should be at least 1 chars long",
			fieldPath:      "spec.secretMounts[0].mountPath",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"Label value is longer than 63 charts": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python39,
					Source:  serverlessv1alpha2.Source{Inline: &serverlessv1alpha2.InlineSource{}},
					Labels: map[string]string{
						strings.Repeat("a", 64): "test",
					},
				},
			},
			expectedErrMsg: "Label value cannot be longer than 63",
			fieldPath:      "spec.labels",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			//WHEN
			err := k8sClient.Create(ctx, tc.fn)

			//THEN
			require.Error(t, err)
			errStatus, ok := err.(*k8serrors.StatusError)
			require.True(t, ok)
			causes := errStatus.Status().Details.Causes
			require.Len(t, causes, 1)
			cause := causes[0]
			assert.Equal(t, tc.expectedCause, cause.Type)
			assert.Equal(t, tc.fieldPath, cause.Field)
			assert.NotEmpty(t, tc.expectedErrMsg, "cause message: %s", cause.Message)
			assert.Contains(t, cause.Message, tc.expectedErrMsg)
		})
	}
}
