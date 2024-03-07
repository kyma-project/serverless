package serverless

import (
	"testing"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_calculateGitImageTag(t *testing.T) {
	tests := []struct {
		name string
		fn   *serverlessv1alpha2.Function
		want string
	}{
		{
			name: "should use runtime",
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
					Runtime: "nodejs18",
				},
			},
			want: "4480d14ea252bf15f18c6632caff283b55beb6f38d5cf8cc43b1b116a151e78d",
		},
		{
			name: "should use runtimeOverride",
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
					Runtime:              "python312",
					RuntimeImageOverride: "nodejs18",
				},
			},
			want: "4480d14ea252bf15f18c6632caff283b55beb6f38d5cf8cc43b1b116a151e78d",
		},
		{
			name: "should use runtime when runtimeOverride is empty",
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
					Runtime:              "nodejs18",
					RuntimeImageOverride: "",
				},
			},
			want: "4480d14ea252bf15f18c6632caff283b55beb6f38d5cf8cc43b1b116a151e78d",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, calculateGitImageTag(tt.fn))
		})
	}
}

func Test_calculateInlineImageTag(t *testing.T) {
	tests := []struct {
		name string
		fn   *serverlessv1alpha2.Function
		want string
	}{
		{
			name: "should use runtime",
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
					Runtime: "nodejs18",
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Commit: "commit",
					Repository: serverlessv1alpha2.Repository{
						BaseDir: "baseDir",
					},
				},
			},
			want: "e6ca45293444d4f6f1b43437f96cd2606842bf7cf9e14a126f52c0b7c216c677",
		},
		{
			name: "should use runtimeOverride",
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
					Runtime:              "python312",
					RuntimeImageOverride: "nodejs18",
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Commit: "commit",
					Repository: serverlessv1alpha2.Repository{
						BaseDir: "baseDir",
					},
				},
			},
			want: "e6ca45293444d4f6f1b43437f96cd2606842bf7cf9e14a126f52c0b7c216c677",
		},
		{
			name: "should use runtime instead of runtimeOverride",
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
					Runtime:              "nodejs18",
					RuntimeImageOverride: "",
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Commit: "commit",
					Repository: serverlessv1alpha2.Repository{
						BaseDir: "baseDir",
					},
				},
			},
			want: "e6ca45293444d4f6f1b43437f96cd2606842bf7cf9e14a126f52c0b7c216c677",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, calculateInlineImageTag(tt.fn))
		})
	}
}
