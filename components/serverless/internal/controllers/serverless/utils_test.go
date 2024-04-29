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
			baseImage: "nodejs18:test",
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
					Runtime: "nodejs20",
				},
			},
			want: "5093e1e9a1b0c94c513bbec23b8291240ba988872353a024d1b0d5b2901d421c",
		},
		{
			name:      "should use runtimeOverride",
			baseImage: "nodejs18:test",
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
					RuntimeImageOverride: "nodejs20",
				},
			},
			want: "5093e1e9a1b0c94c513bbec23b8291240ba988872353a024d1b0d5b2901d421c",
		},
		{
			name:      "should use runtime when runtimeOverride is empty",
			baseImage: "nodejs18:test",
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
			baseImage: "nodejs18:test",
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
					Runtime: "nodejs20",
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Commit: "commit",
					Repository: serverlessv1alpha2.Repository{
						BaseDir: "baseDir",
					},
				},
			},
			want: "3773f4e6b48ccb71c2a43f4a7cbbeb98333701d823d8ff5917d8f7373c7abbcd",
		},
		{
			name:      "should use runtimeOverride",
			baseImage: "nodejs18:test",
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
					RuntimeImageOverride: "nodejs20",
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Commit: "commit",
					Repository: serverlessv1alpha2.Repository{
						BaseDir: "baseDir",
					},
				},
			},
			want: "3773f4e6b48ccb71c2a43f4a7cbbeb98333701d823d8ff5917d8f7373c7abbcd",
		},
		{
			name:      "should use runtime instead of runtimeOverride",
			baseImage: "nodejs18:test",
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
					RuntimeImageOverride: "",
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Commit: "commit",
					Repository: serverlessv1alpha2.Repository{
						BaseDir: "baseDir",
					},
				},
			},
			want: "3773f4e6b48ccb71c2a43f4a7cbbeb98333701d823d8ff5917d8f7373c7abbcd",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, calculateInlineImageTag(tt.fn, tt.baseImage))
		})
	}
}
