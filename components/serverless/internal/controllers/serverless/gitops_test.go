package serverless

import (
	"testing"

	"github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
	"github.com/onsi/gomega"
)

func Test_isOnSourceChange(t *testing.T) {
	testCases := []struct {
		desc           string
		fn             v1alpha2.Function
		revision       string
		expectedResult bool
	}{
		{
			desc: "new function",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{},
					},
					Runtime: v1alpha2.NodeJs20,
				},
			},
			expectedResult: true,
		},
		{
			desc: "new function fixed on commit",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{
							Repository: v1alpha2.Repository{
								Reference: "1",
							},
						},
					},
					Runtime: v1alpha2.NodeJs20,
				},
			},
			expectedResult: true,
		},
		{
			desc: "new function follow head",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{
							Repository: v1alpha2.Repository{
								Reference: "1",
							},
						},
					},
					Runtime: v1alpha2.NodeJs20,
				},
			},
			expectedResult: true,
		},
		{
			desc: "function did not change",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{
							Repository: v1alpha2.Repository{
								Reference: "1",
							},
						},
					},
					Runtime: v1alpha2.NodeJs20,
				},
				Status: v1alpha2.FunctionStatus{
					Repository: v1alpha2.Repository{
						Reference: "1",
					},
					Commit:  "1",
					Runtime: v1alpha2.NodeJs20,
				},
			},
			revision:       "1",
			expectedResult: false,
		},
		{
			desc: "function change fixed revision",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{
							Repository: v1alpha2.Repository{
								Reference: "2",
							},
						},
					},
					Runtime: v1alpha2.NodeJs20,
				},
				Status: v1alpha2.FunctionStatus{
					Repository: v1alpha2.Repository{
						Reference: "1",
					},
				},
			},
			expectedResult: true,
		},
		{
			desc: "function change",
			fn: v1alpha2.Function{
				Status: v1alpha2.FunctionStatus{
					Repository: v1alpha2.Repository{
						Reference: "1",
					},
				},
			},
			revision:       "2",
			expectedResult: true,
		},
		{
			desc: "function change base dir",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{
							Repository: v1alpha2.Repository{
								Reference: "2",
								BaseDir:   "base_dir",
							},
						},
					},
				},
				Status: v1alpha2.FunctionStatus{
					Repository: v1alpha2.Repository{
						Reference: "2",
					},
				},
			},
			expectedResult: true,
		},
		{
			desc: "function change branch",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{
							Repository: v1alpha2.Repository{
								Reference: "branch",
							},
						},
					},
				},
				Status: v1alpha2.FunctionStatus{
					Repository: v1alpha2.Repository{
						Reference: "2",
					},
				},
			},
			expectedResult: true,
		},
		{
			desc: "function change dockerfile",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{
							Repository: v1alpha2.Repository{
								Reference: "2",
							},
						},
					},
					Runtime: v1alpha2.NodeJs20,
				},
				Status: v1alpha2.FunctionStatus{
					Repository: v1alpha2.Repository{
						Reference: "2",
					},
				},
			},
			expectedResult: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			s := systemState{
				instance: tC.fn,
			}
			actual := s.gitFnSrcChanged(tC.revision)
			g.Expect(actual).To(gomega.Equal(tC.expectedResult))
		})
	}
}
