package v1alpha2

import (
	"testing"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFunctionSpec_validateResources(t *testing.T) {
	for testName, testData := range map[string]struct {
		givenFunc              Function
		expectedError          gomega.OmegaMatcher
		specifiedExpectedError gomega.OmegaMatcher
	}{
		"Should return errors on empty function": {
			givenFunc:     Function{},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring(
					"metadata.name",
				),
				gomega.ContainSubstring(
					"metadata.namespace",
				),
			),
		},
		"Should be ok": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: Source{
						Inline: &InlineSource{
							Source: "test-source",
						},
					},
					Runtime: NodeJs18,
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
						},
						Build: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("300m"),
									corev1.ResourceMemory: resource.MustParse("300Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("200Mi"),
								},
							},
						},
					},
				},
			},
			expectedError: gomega.BeNil(),
		},
		"Should validate all fields without error": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: Source{
						Inline: &InlineSource{
							Source:       "test-source",
							Dependencies: " { test }     \t\n",
						},
					},
					Runtime: NodeJs18,
					Env: []corev1.EnvVar{
						{
							Name:  "test",
							Value: "test",
						},
						{
							Name:  "config",
							Value: "test",
						},
					},
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
						},
						Build: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("300m"),
									corev1.ResourceMemory: resource.MustParse("300Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("200Mi"),
								},
							},
						},
					},
					SecretMounts: []SecretMount{
						{
							SecretName: "secret-name-1",
							MountPath:  "/mount/path/1",
						},
						{
							SecretName: "secret-name-2",
							MountPath:  "/mount/path/2",
						},
					},
					Labels: map[string]string{
						"label-1": "label-1-value",
						"label-2": "label-2-value",
					},
					Annotations: map[string]string{
						"annotation-1": "annotation-1-value",
						"annotation-2": "annotation-2-value",
					},
				},
			},
			expectedError: gomega.BeNil(),
		},
		"should be OK for git sourceType": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: Source{
						GitRepository: &GitRepositorySource{
							URL: "test-source",
							Repository: Repository{
								BaseDir:   "/",
								Reference: "test-me",
							},
						},
					},
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
						},
						Build: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("400m"),
									corev1.ResourceMemory: resource.MustParse("400Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("300m"),
									corev1.ResourceMemory: resource.MustParse("300Mi"),
								},
							},
						},
					},
					Runtime: NodeJs18,
				},
			},
			expectedError: gomega.BeNil(),
		},
		"Should return errors OK if reference and baseDir is missing": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: Source{
						GitRepository: &GitRepositorySource{
							URL: "testme",
						},
					},
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
						},
						Build: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("200Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("200Mi"),
								},
							},
						},
					},
					Runtime: NodeJs18,
				},
			},
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring("spec.source.gitRepository.reference"),
				gomega.ContainSubstring("spec.source.gitRepository.baseDir"),
			),
			expectedError: gomega.HaveOccurred(),
		},
		"Should validate without error Resources and Profile occurring at once in ResourceConfiguration.Function/Build": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: Source{
						Inline: &InlineSource{
							Source:       "test-source",
							Dependencies: " { test }",
						},
					},
					Runtime: NodeJs18,
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Profile: "function-profile",
							Resources: &corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
						},
						Build: &ResourceRequirements{
							Profile: "build-profile",
							Resources: &corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("300m"),
									corev1.ResourceMemory: resource.MustParse("300Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("200Mi"),
								},
							},
						},
					},
				},
			},
			expectedError: gomega.BeNil(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			tn := testName
			t.Log(tn)
			// given
			g := gomega.NewWithT(t)
			config := fixValidationConfig()

			// when
			errs := testData.givenFunc.Validate(config)
			// then
			g.Expect(errs).To(testData.expectedError)
			if testData.specifiedExpectedError != nil {
				g.Expect(errs.Error()).To(testData.specifiedExpectedError)
			}
		})
	}
}

func TestFunctionSpec_validateGitRepoURL(t *testing.T) {

	tests := []struct {
		name    string
		spec    FunctionSpec
		wantErr bool
	}{
		{
			name: "Invalid http",
			spec: FunctionSpec{
				Source: Source{
					GitRepository: &GitRepositorySource{
						URL: "github.com/kyma-project/kyma.git",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Valid http",
			spec: FunctionSpec{
				Source: Source{
					GitRepository: &GitRepositorySource{
						URL: "https://github.com/kyma-project/kyma.git",
					},
				},
			},
		},
		{
			name: "Invalid ssh",
			spec: FunctionSpec{
				Source: Source{
					GitRepository: &GitRepositorySource{
						URL: "g0t@github.com:kyma-project/kyma.git",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Valid ssh",
			spec: FunctionSpec{
				Source: Source{
					GitRepository: &GitRepositorySource{
						URL: "git@github.com:kyma-project/kyma.git",
					},
				},
			},
		},
		{
			name: "Valid ssh without .git extension",
			spec: FunctionSpec{
				Source: Source{
					GitRepository: &GitRepositorySource{
						URL: "git@github.com:kyma-project/kyma",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := tt.spec.validateGitRepoURL(&ValidationConfig{}); (err != nil) != tt.wantErr {
				t.Errorf("FunctionSpec.validateGitRepoURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func fixValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		ReservedEnvs: []string{"K_CONFIGURATION"},
		Function: MinFunctionValues{
			Resources: MinFunctionResourcesValues{
				MinRequestCPU:    "10m",
				MinRequestMemory: "16Mi",
			},
		},
		BuildJob: MinBuildJobValues{
			Resources: MinBuildJobResourcesValues{
				MinRequestCPU:    "200m",
				MinRequestMemory: "200Mi",
			},
		},
	}
}
