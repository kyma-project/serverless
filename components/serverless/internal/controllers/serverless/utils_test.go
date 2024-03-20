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
					Runtime: "nodejs20",
				},
			},
			want: "5e62e84b27afdcf23e9ea682a8ce44b693c4a3258e5b26bd038c60cd41eb60ee",
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
					Runtime:              "nodejs18",
					RuntimeImageOverride: "nodejs20",
				},
			},
			want: "5e62e84b27afdcf23e9ea682a8ce44b693c4a3258e5b26bd038c60cd41eb60ee",
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
					Runtime:              "nodejs20",
					RuntimeImageOverride: "",
				},
			},
			want: "5e62e84b27afdcf23e9ea682a8ce44b693c4a3258e5b26bd038c60cd41eb60ee",
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
					Runtime: "nodejs20",
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Commit: "commit",
					Repository: serverlessv1alpha2.Repository{
						BaseDir: "baseDir",
					},
				},
			},
			want: "9f131e00ad3c6cfc5ca36f27df299eeeb2b08bcc4328782e79b69440b1b7aa2b",
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
			want: "9f131e00ad3c6cfc5ca36f27df299eeeb2b08bcc4328782e79b69440b1b7aa2b",
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
			want: "9f131e00ad3c6cfc5ca36f27df299eeeb2b08bcc4328782e79b69440b1b7aa2b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, calculateInlineImageTag(tt.fn))
		})
	}
}
