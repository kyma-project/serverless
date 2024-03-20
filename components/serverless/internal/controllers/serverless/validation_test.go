package serverless

import (
	"context"
	"testing"

	"github.com/kyma-project/serverless/components/serverless/internal/controllers/serverless/automock"
	serverlessResource "github.com/kyma-project/serverless/components/serverless/internal/resource"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	minResourcesCfg = ResourceConfig{
		Function: FunctionResourceConfig{
			Resources: Resources{
				MinRequestedCPU:    Quantity{resource.MustParse("10m")},
				MinRequestedMemory: Quantity{resource.MustParse("10Mi")},
			},
		},
		BuildJob: BuildJobResourceConfig{
			Resources: Resources{
				MinRequestedCPU:    Quantity{resource.MustParse("20m")},
				MinRequestedMemory: Quantity{resource.MustParse("20Mi")},
			},
		},
	}
)

func TestValidation_Invalid(t *testing.T) {
	//GIVEN
	ctx := context.TODO()

	k8sClient := fake.NewClientBuilder().WithStatusSubresource(&serverlessv1alpha2.Function{}).Build()
	require.NoError(t, serverlessv1alpha2.AddToScheme(scheme.Scheme))
	resourceClient := serverlessResource.New(k8sClient, scheme.Scheme)

	statsCollector := &automock.StatsCollector{}
	statsCollector.On("UpdateReconcileStats", mock.Anything, mock.Anything).Return()

	testCases := map[string]struct {
		fn              serverlessv1alpha2.Function
		expectedCondMsg string
	}{
		"Function requests cpu are bigger than limits": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("150m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Function limits cpu(120m) should be higher than requests cpu(150m)",
		},
		"Function requests memory are bigger than limits": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("150Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Function limits memory(120Mi) should be higher than requests memory(150Mi)",
		},
		"Function requests cpu are smaller than minimum value": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("5m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Function request cpu(5m) should be higher than minimal value (10m)",
		},
		"Function requests memory are smaller than minimum value": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("5Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Function request memory(5Mi) should be higher than minimal value (10Mi)",
		},
		"Function limits cpu are smaller than minimum without requests": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("2m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
						}},
					},
				},
			},
			expectedCondMsg: "Function limits cpu(2m) should be higher than minimal value (10m)",
		},
		"Function limits memory are smaller than minimum without requests": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("20m"),
								corev1.ResourceMemory: resource.MustParse("2Mi"),
							},
						}},
					},
				},
			},
			expectedCondMsg: "Function limits memory(2Mi) should be higher than minimal value (10Mi)",
		},
		//Build validation
		"Build requests cpu are bigger than limits": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("150m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Build limits cpu(120m) should be higher than requests cpu(150m)",
		},
		"Build requests memory are bigger than limits": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("150Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Build limits memory(120Mi) should be higher than requests memory(150Mi)",
		},
		"Build requests cpu are smaller than minimum value": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("5m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Build request cpu(5m) should be higher than minimal value (20m)",
		},
		"Build requests memory are smaller than minimum value": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("5Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Build request memory(5Mi) should be higher than minimal value (20Mi)",
		},
		"Build limits cpu are smaller than minimum without requests": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("2m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
						}},
					},
				},
			},
			expectedCondMsg: "Build limits cpu(2m) should be higher than minimal value (20m)",
		},
		"Build limits memory are smaller than minimum without requests": {
			fn: serverlessv1alpha2.Function{ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("30m"),
								corev1.ResourceMemory: resource.MustParse("2Mi"),
							},
						}},
					},
				},
			},
			expectedCondMsg: "Build limits memory(2Mi) should be higher than minimal value (20Mi)",
		},
		"Invalid env": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Env: []corev1.EnvVar{
						{Name: "1ENV"},
						{Name: "2ENV"},
					},
				},
			},
			expectedCondMsg: "spec.env: 1ENV. Err: a valid environment variable name must consist of alphabetic characters, digits, '_', '-', or '.', and must not start with a digit (e.g. 'my.env-name',  or 'MY_ENV.NAME',  or 'MyEnvName1', regex used for validation is '[-._a-zA-Z][-._a-zA-Z0-9]*')",
		},
		"Invalid secretMount name": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					SecretMounts: []serverlessv1alpha2.SecretMount{
						{
							SecretName: "secret-name-1",
							MountPath:  "/mount/path/1",
						},
						{
							SecretName: "invalid secret name - not DNS subdomain name as defined in RFC 1123",
							MountPath:  "/mount/path/2",
						},
					},
				},
			},
			expectedCondMsg: "invalid spec.secretMounts: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
		},
		"Non unique secretMount name": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					SecretMounts: []serverlessv1alpha2.SecretMount{
						{
							SecretName: "secret-name-1",
							MountPath:  "/mount/path/1",
						},
						{
							SecretName: "non-unique-secret-name",
							MountPath:  "/mount/path/2",
						},
						{
							SecretName: "non-unique-secret-name",
							MountPath:  "/mount/path/3",
						},
					},
				},
			},
			expectedCondMsg: "invalid spec.secretMounts: [secretNames should be unique]",
		},
		"Improper dependencies for JS": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs20,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "source code",
							Dependencies: "invalid dependencies",
						},
					},
				},
			},
			expectedCondMsg: "invalid source.inline.dependencies value: deps should start with '{' and end with '}'",
		},
		"Invalid label name": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Labels: map[string]string{
						".invalid-name": "value",
					},
				},
			},
			expectedCondMsg: "spec.labels: Invalid value: \".invalid-name\": name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')",
		},
		"Invalid label value": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Labels: map[string]string{
						"name": ".invalid-value",
					},
				},
			},
			expectedCondMsg: "spec.labels: Invalid value: \".invalid-value\": a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')",
		},
		"Invalid annotation name": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Annotations: map[string]string{
						".invalid-name": "value",
					},
				},
			},
			expectedCondMsg: "spec.annotations: Invalid value: \".invalid-name\": name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')",
		},
		"Invalid Git repository url": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs20,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: "abc",
						},
					},
				},
			},
			expectedCondMsg: "source.gitRepository.URL: parse \"abc\": invalid URI for request",
		},
		"Invalid Git repository http url": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs20,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: "github.com/kyma-project/kyma.git",
						},
					},
				},
			},
			expectedCondMsg: "source.gitRepository.URL: parse \"github.com/kyma-project/kyma.git\": invalid URI for request",
		},
		"Invalid Git repository ssh url": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs20,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: "g0t@github.com:kyma-project/kyma.git",
						},
					},
				},
			},
			expectedCondMsg: "source.gitRepository.URL: parse \"g0t@github.com:kyma-project/kyma.git\": invalid URI for request",
		},
	}

	//WHEN
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, resourceClient.Create(ctx, &tc.fn))
			s := &systemState{instance: tc.fn}
			r := &reconciler{out: out{result: controllerruntime.Result{Requeue: true}},
				k8s: k8s{client: resourceClient, recorder: record.NewFakeRecorder(100), statsCollector: statsCollector},
				cfg: cfg{fn: FunctionConfig{ResourceConfig: minResourcesCfg}}}

			//WHEN
			nextFn, err := stateFnValidateFunction(ctx, r, s)
			require.NoError(t, err)
			_, err = nextFn(context.TODO(), r, s)

			//THEN
			require.NoError(t, err)
			updatedFn := serverlessv1alpha2.Function{}
			require.NoError(t, resourceClient.Get(ctx, ctrlclient.ObjectKey{Name: tc.fn.Name}, &updatedFn))
			cond := getCondition(updatedFn.Status.Conditions, serverlessv1alpha2.ConditionConfigurationReady)
			assert.Equal(t, serverlessv1alpha2.ConditionReasonFunctionSpec, cond.Reason)
			assert.Equal(t, corev1.ConditionFalse, cond.Status)
			assert.NotEmpty(t, tc.expectedCondMsg, "expected message shouldn't be empty")
			assert.Equal(t, tc.expectedCondMsg, cond.Message)
			assert.False(t, r.result.Requeue)
		})
	}
}

func TestValidation_Valid(t *testing.T) {
	//GIVEN
	ctx := context.TODO()

	k8sClient := fake.NewClientBuilder().Build()
	require.NoError(t, serverlessv1alpha2.AddToScheme(scheme.Scheme))
	resourceClient := serverlessResource.New(k8sClient, scheme.Scheme)

	statsCollector := &automock.StatsCollector{}
	statsCollector.On("UpdateReconcileStats", mock.Anything, mock.Anything).Return()

	testCases := map[string]struct {
		fn serverlessv1alpha2.Function
	}{
		"Valid function": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("100Mi"),
							}}},
					},
					Env: []corev1.EnvVar{
						{Name: "_CORRECT_ENV"},
						{Name: "ANOTHER_CORRECT_ENV"},
					},
					SecretMounts: []serverlessv1alpha2.SecretMount{
						{
							SecretName: "secret-name",
							MountPath:  "mount-path",
						},
					},
					Labels: map[string]string{
						"name1": "value1",
						"name2": "value2",
						"name3": "",
					},
					Annotations: map[string]string{
						"name1": "value1",
						"name2": "value2",
					},
				},
			},
		},
		"Dependencies for JS": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs20,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "source code",
							Dependencies: "{valid javascript dependencies}",
						},
					},
				},
			},
		},
		"Dependencies for Python": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.Python312,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "source code",
							Dependencies: "valid python dependencies",
						},
					},
				},
			},
		},
		"Empty dependencies": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs20,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source: "source code",
						},
					},
				},
			},
		},
		"Git repository URL SSH": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs20,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: "git@github.com:kyma-project/serverless.git",
						},
					},
				},
			},
		},
		"Git repository URL HTTPS": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs20,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: "https://github.com/kyma-project/serverless.git",
						},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, resourceClient.Create(ctx, &tc.fn))
			s := &systemState{instance: tc.fn}
			r := &reconciler{out: out{result: controllerruntime.Result{Requeue: true}},
				k8s: k8s{client: resourceClient, recorder: record.NewFakeRecorder(100), statsCollector: statsCollector},
				cfg: cfg{fn: FunctionConfig{ResourceConfig: minResourcesCfg}}}

			//WHEN
			nextFn, err := stateFnValidateFunction(ctx, r, s)
			require.NoError(t, err)
			_, err = nextFn(context.TODO(), r, s)

			//THEN
			require.NoError(t, err)
			updatedFn := serverlessv1alpha2.Function{}
			require.NoError(t, resourceClient.Get(ctx, ctrlclient.ObjectKey{Name: tc.fn.Name}, &updatedFn))
			assert.True(t, r.result.Requeue)

		})
	}
}
