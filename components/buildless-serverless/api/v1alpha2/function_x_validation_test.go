package v1alpha2_test

import (
	"context"
	"strings"
	"testing"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/testenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// GIVEN
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
				},
			},
		},
		"Valid inline source with Python runtime": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "print('Hello, World!')",
							Dependencies: "requests==2.25.1",
						},
					},
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
				},
			},
		},
		"allowed runtime: nodejs20": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs20,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
				},
			},
		},
		"allowed runtime: nodejs22": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs22,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
				},
			},
		},
		"allowed runtime: python312": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
				},
			},
		},
		"allowed envs": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Env: []corev1.EnvVar{{Name: "TEST_ENV"}, {Name: "MY_ENV"}},
				},
			},
		},
		"gitRepository used as source": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "base-dir",
								Reference: "ref",
							},
						},
					},
				},
			},
		},
		"Git source has auth with key Type": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "dir",
								Reference: "ref",
							},
							Auth: &serverlessv1alpha2.RepositoryAuth{
								Type:       "key",
								SecretName: "secret",
							},
						},
					},
				},
			},
		},
		"Git source has auth with basic Type": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "dir",
								Reference: "ref",
							},
							Auth: &serverlessv1alpha2.RepositoryAuth{
								Type:       "basic",
								SecretName: "secret",
							},
						},
					},
				},
			},
		},
		"valid Git source with Node.js runtime": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs20,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: "https://github.com/example/repo.git",
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "src",
								Reference: "main",
							},
						},
					},
				},
			},
		},
		"valid Git source with authentication": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs22,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: "git@github.com:example/repo.git",
							Auth: &serverlessv1alpha2.RepositoryAuth{
								Type:       serverlessv1alpha2.RepositoryAuthSSHKey,
								SecretName: "git-ssh-secret",
							},
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "app",
								Reference: "develop",
							},
						},
					},
				},
			},
		},
		"label value length is 63": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "abc"}},
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
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "abc"}},
					SecretMounts: []serverlessv1alpha2.SecretMount{{MountPath: "/path", SecretName: "secret-name"}},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// WHEN
			err := k8sClient.Create(ctx, tc.fn)

			// THEN
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			expectedErrMsg: "Invalid value: Use profile or resources",
			fieldPath:      "spec.resourceConfiguration.function",
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			expectedErrMsg: "Invalid value: Invalid profile, please use one of: ['XS','S','M','L','XL']",
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			fieldPath:      "spec.labels",
			expectedErrMsg: "Invalid value: Labels has key starting with serverless.kyma-project.io/ which is not allowed",
		},
		"labels use restricted domain": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Labels: map[string]string{
						"serverless.kyma-project.io/abc": "labelValue",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			fieldPath:      "spec.labels",
			expectedErrMsg: "Invalid value: Labels has key starting with serverless.kyma-project.io/ which is not allowed",
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			fieldPath:      "spec.labels",
			expectedErrMsg: "Invalid value: Labels has key starting with serverless.kyma-project.io/ which is not allowed",
		},
		"annotations use exact restricted domain": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Annotations: map[string]string{
						"serverless.kyma-project.io/": "labelValue",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			fieldPath:      "spec.annotations",
			expectedErrMsg: "Invalid value: Annotations has key starting with serverless.kyma-project.io/ which is not allowed",
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
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			fieldPath:      "spec.annotations",
			expectedErrMsg: "Invalid value: Annotations has key starting with serverless.kyma-project.io/ which is not allowed",
		},
		"annotations use restricted domain": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Annotations: map[string]string{
						"serverless.kyma-project.io/abc": "labelValue",
					},
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			fieldPath:      "spec.annotations",
			expectedErrMsg: "Invalid value: Annotations has key starting with serverless.kyma-project.io/ which is not allowed",
		},
		"reserved env: FUNC_RUNTIME": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
					Env: []corev1.EnvVar{
						{Name: "TEST2"},
						{Name: "FUNC_RUNTIME"},
						{Name: "TEST"},
					},
				},
			},
			expectedErrMsg: "Invalid value: Following envs are reserved and cannot be used: ['FUNC_RUNTIME','FUNC_HANDLER','FUNC_PORT','FUNC_HANDLER_SOURCE','FUNC_HANDLER_DEPENDENCIES','MOD_NAME','NODE_PATH','PYTHONPATH']",
			fieldPath:      "spec.env",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"reserved env: FUNC_HANDLER": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
					Env: []corev1.EnvVar{
						{Name: "TEST2"},
						{Name: "FUNC_HANDLER"},
						{Name: "TEST"},
					},
				},
			},
			expectedErrMsg: "Invalid value: Following envs are reserved and cannot be used: ['FUNC_RUNTIME','FUNC_HANDLER','FUNC_PORT','FUNC_HANDLER_SOURCE','FUNC_HANDLER_DEPENDENCIES','MOD_NAME','NODE_PATH','PYTHONPATH']",
			fieldPath:      "spec.env",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"reserved env: FUNC_PORT": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
					Env: []corev1.EnvVar{
						{Name: "TEST2"},
						{Name: "FUNC_PORT"},
						{Name: "TEST"},
					},
				},
			},
			expectedErrMsg: "Invalid value: Following envs are reserved and cannot be used: ['FUNC_RUNTIME','FUNC_HANDLER','FUNC_PORT','FUNC_HANDLER_SOURCE','FUNC_HANDLER_DEPENDENCIES','MOD_NAME','NODE_PATH','PYTHONPATH']",
			fieldPath:      "spec.env",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"reserved env: MOD_NAME": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
					Env: []corev1.EnvVar{
						{Name: "TEST2"},
						{Name: "MOD_NAME"},
						{Name: "TEST"},
					},
				},
			},
			expectedErrMsg: "Invalid value: Following envs are reserved and cannot be used: ['FUNC_RUNTIME','FUNC_HANDLER','FUNC_PORT','FUNC_HANDLER_SOURCE','FUNC_HANDLER_DEPENDENCIES','MOD_NAME','NODE_PATH','PYTHONPATH']",
			fieldPath:      "spec.env",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"reserved env: NODE_PATH": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
					Env: []corev1.EnvVar{
						{Name: "TEST2"},
						{Name: "NODE_PATH"},
						{Name: "TEST"},
					},
				},
			},
			expectedErrMsg: "Invalid value: Following envs are reserved and cannot be used: ['FUNC_RUNTIME','FUNC_HANDLER','FUNC_PORT','FUNC_HANDLER_SOURCE','FUNC_HANDLER_DEPENDENCIES','MOD_NAME','NODE_PATH','PYTHONPATH']",
			fieldPath:      "spec.env",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"reserved env: PYTHONPATH": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Python312,
					Env: []corev1.EnvVar{
						{Name: "TEST2"},
						{Name: "PYTHONPATH"},
						{Name: "TEST"},
					},
				},
			},
			expectedErrMsg: "Invalid value: Following envs are reserved and cannot be used: ['FUNC_RUNTIME','FUNC_HANDLER','FUNC_PORT','FUNC_HANDLER_SOURCE','FUNC_HANDLER_DEPENDENCIES','MOD_NAME','NODE_PATH','PYTHONPATH']",
			fieldPath:      "spec.env",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"GitRepository and Inline source used together": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						GitRepository: nil,
						Inline:        nil,
					},
				},
			},
			expectedErrMsg: "Invalid value: Use GitRepository or Inline source",
			fieldPath:      "spec.source",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"Neither GitRepository nor Inline source was used": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source:  serverlessv1alpha2.Source{},
				},
			},
			expectedErrMsg: "Invalid value: Use GitRepository or Inline source",
			fieldPath:      "spec.source",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"Secret Mount name is empty": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source:  serverlessv1alpha2.Source{Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					SecretMounts: []serverlessv1alpha2.SecretMount{
						{SecretName: "", MountPath: "/path"},
					},
				},
			},
			expectedErrMsg: "Invalid value: \"\": spec.secretMounts[0].secretName in body should be at least 1 chars long",
			fieldPath:      "spec.secretMounts[0].secretName",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"Secret Mount path is empty": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source:  serverlessv1alpha2.Source{Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					SecretMounts: []serverlessv1alpha2.SecretMount{
						{SecretName: "my-secret", MountPath: ""},
					},
				},
			},
			expectedErrMsg: "Invalid value: \"\": spec.secretMounts[0].mountPath in body should be at least 1 chars long",
			fieldPath:      "spec.secretMounts[0].mountPath",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"Label value is longer than 63 charts": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source:  serverlessv1alpha2.Source{Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Labels: map[string]string{
						strings.Repeat("a", 64): "test",
					},
				},
			},
			expectedErrMsg: "Invalid value: Label value cannot be longer than 63",
			fieldPath:      "spec.labels",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"Inline source is empty": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{},
					},
				},
			},
			expectedErrMsg: "Invalid value: \"\": spec.source.inline.source in body should be at least 1 chars long",
			fieldPath:      "spec.source.inline.source",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"Git source has empty BaseDir": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "   ",
								Reference: "ref",
							},
						},
					},
				},
			},
			expectedErrMsg: "Invalid value: BaseDir is required and cannot be empty",
			fieldPath:      "spec.source.gitRepository",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"Git source has empty Reference": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "dir",
								Reference: "   ",
							},
						},
					},
				},
			},
			expectedErrMsg: "Invalid value: Reference is required and cannot be empty",
			fieldPath:      "spec.source.gitRepository",
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
		},
		"Git source has empty SecretName": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "dir",
								Reference: "ref",
							},
							Auth: &serverlessv1alpha2.RepositoryAuth{
								Type:       "key",
								SecretName: "  ",
							},
						},
					},
				},
			},
			expectedErrMsg: "Invalid value: \"  \": SecretName is required and cannot be empty",
			fieldPath:      "spec.source.gitRepository.auth.secretName",
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
			assert.Equal(t, tc.expectedErrMsg, cause.Message)
		})
	}
}

// in some cases k8s 1.30+ returns two reasons
func Test_XKubernetesValidations_InvalidMultipleCauses(t *testing.T) {
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
		"disallowed runtime: custom": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{Source: "a"}},
					Runtime: serverlessv1alpha2.Runtime("custom"),
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueNotSupported,
			fieldPath:      "spec.runtime",
			expectedErrMsg: "Unsupported value: \"custom\": supported values: \"nodejs20\", \"nodejs22\", \"python312\"",
		},
		"Git source auth has incorrect Type": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "dir",
								Reference: "ref",
							},
							Auth: &serverlessv1alpha2.RepositoryAuth{
								Type:       "custom",
								SecretName: "secret",
							},
						},
					},
				},
			},
			expectedErrMsg: "Unsupported value: \"custom\": supported values: \"basic\", \"key\"",
			fieldPath:      "spec.source.gitRepository.auth.type",
			expectedCause:  metav1.CauseTypeFieldValueNotSupported,
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
			require.Len(t, causes, 2)
			causeFound := false
			for _, cause := range causes {
				if cause.Type == tc.expectedCause && cause.Field == tc.fieldPath {
					causeFound = true
					assert.Equal(t, tc.expectedCause, cause.Type)
					assert.Equal(t, tc.fieldPath, cause.Field)
					assert.NotEmpty(t, tc.expectedErrMsg, "cause message: %s", cause.Message)
					assert.Equal(t, tc.expectedErrMsg, cause.Message)
				}
			}
			assert.True(t, causeFound)
		})
	}
}
