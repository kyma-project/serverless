package serverless

import (
	"testing"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_calculateGitImageTag(t *testing.T) {
	tests := []struct {
		name      string
		fn        *serverlessv1alpha2.Function
		baseImage string
		want      string
	}{
		{
			name:      "should use runtime",
			baseImage: "nodejs22:test",
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					UID: "fn-uuid",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "fn-source",
							Dependencies: "",
						},
					},
					Runtime: "nodejs22",
				},
			},
			want: "da01b5198b936f587379f6311ffc1a557e2d680c2fc88a61e6777fbc1d3dac48",
		},
		{
			name:      "should use runtimeOverride",
			baseImage: "nodejs22:test",
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					UID: "fn-uuid",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "fn-source",
							Dependencies: "",
						},
					},
					Runtime:              "nodejs22",
					RuntimeImageOverride: "nodejs22",
				},
			},
			want: "da01b5198b936f587379f6311ffc1a557e2d680c2fc88a61e6777fbc1d3dac48",
		},
		{
			name:      "should use runtime when runtimeOverride is empty",
			baseImage: "nodejs22:test",
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					UID: "fn-uuid",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "fn-source",
							Dependencies: "",
						},
					},
					Runtime:              "nodejs22",
					RuntimeImageOverride: "",
				},
			},
			want: "da01b5198b936f587379f6311ffc1a557e2d680c2fc88a61e6777fbc1d3dac48",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, calculateGitImageTag(tt.fn, tt.baseImage))
		})
	}
}

func Test_calculateInlineImageTag(t *testing.T) {
	tests := []struct {
		name      string
		fn        *serverlessv1alpha2.Function
		baseImage string
		want      string
	}{
		{
			name:      "should use runtime",
			baseImage: "nodejs22:test",
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					UID: "fn-uuid",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "fn-source",
							Dependencies: "",
						},
					},
					Runtime: "nodejs22",
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Commit: "commit",
					Repository: serverlessv1alpha2.Repository{
						BaseDir: "baseDir",
					},
				},
			},
			want: "7bcecf54edf9aecbc68fd10db1349f29866b6d0157f744841371290977f09dcb",
		},
		{
			name:      "should use runtimeOverride",
			baseImage: "nodejs22:test",
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					UID: "fn-uuid",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "fn-source",
							Dependencies: "",
						},
					},
					Runtime:              "nodejs22",
					RuntimeImageOverride: "nodejs22",
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Commit: "commit",
					Repository: serverlessv1alpha2.Repository{
						BaseDir: "baseDir",
					},
				},
			},
			want: "7bcecf54edf9aecbc68fd10db1349f29866b6d0157f744841371290977f09dcb",
		},
		{
			name:      "should use runtime instead of runtimeOverride",
			baseImage: "nodejs22:test",
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					UID: "fn-uuid",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "fn-source",
							Dependencies: "",
						},
					},
					Runtime:              "nodejs22",
					RuntimeImageOverride: "",
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Commit: "commit",
					Repository: serverlessv1alpha2.Repository{
						BaseDir: "baseDir",
					},
				},
			},
			want: "7bcecf54edf9aecbc68fd10db1349f29866b6d0157f744841371290977f09dcb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, calculateInlineImageTag(tt.fn, tt.baseImage))
		})
	}
}

// functions only used in tests
func getConditionReason(conditions []serverlessv1alpha2.Condition, conditionType serverlessv1alpha2.ConditionType) serverlessv1alpha2.ConditionReason {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Reason
		}
	}

	return ""
}

func getCondition(conditions []serverlessv1alpha2.Condition, conditionType serverlessv1alpha2.ConditionType) serverlessv1alpha2.Condition {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition
		}
	}

	return serverlessv1alpha2.Condition{}
}
