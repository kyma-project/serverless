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
			baseImage: "nodejs20:test",
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
			want: "5093e1e9a1b0c94c513bbec23b8291240ba988872353a024d1b0d5b2901d421c",
		},
		{
			name:      "should use runtimeOverride",
			baseImage: "nodejs20:test",
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
					Runtime:              "nodejs20",
					RuntimeImageOverride: "nodejs22",
				},
			},
			want: "5093e1e9a1b0c94c513bbec23b8291240ba988872353a024d1b0d5b2901d421c",
		},
		{
			name:      "should use runtime when runtimeOverride is empty",
			baseImage: "nodejs20:test",
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
			want: "5093e1e9a1b0c94c513bbec23b8291240ba988872353a024d1b0d5b2901d421c",
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
			baseImage: "nodejs20:test",
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
			want: "22c34628c8c23fa46e86769359820d6809390f4064b51ced1f9529d4855af5a4",
		},
		{
			name:      "should use runtimeOverride",
			baseImage: "nodejs20:test",
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
					Runtime:              "nodejs20",
					RuntimeImageOverride: "nodejs22",
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Commit: "commit",
					Repository: serverlessv1alpha2.Repository{
						BaseDir: "baseDir",
					},
				},
			},
			want: "22c34628c8c23fa46e86769359820d6809390f4064b51ced1f9529d4855af5a4",
		},
		{
			name:      "should use runtime instead of runtimeOverride",
			baseImage: "nodejs20:test",
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
			want: "22c34628c8c23fa46e86769359820d6809390f4064b51ced1f9529d4855af5a4",
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
