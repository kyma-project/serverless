package runtime_test

import (
	"testing"

	"github.com/kyma-project/serverless/components/serverless/internal/controllers/serverless/runtime"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func TestGetRuntimeConfig(t *testing.T) {
	for testName, testData := range map[string]struct {
		name    string
		runtime serverlessv1alpha2.Runtime
		want    runtime.Config
	}{
		"python39": {
			name:    "python39",
			runtime: serverlessv1alpha2.Python39,
			want: runtime.Config{
				Runtime:                 serverlessv1alpha2.Python39,
				DependencyFile:          "requirements.txt",
				FunctionFile:            "handler.py",
				DockerfileConfigMapName: "dockerfile-python39",
				RuntimeEnvs: []corev1.EnvVar{{Name: "PYTHONPATH", Value: "$(KUBELESS_INSTALL_VOLUME)/lib.python3.9/site-packages:$(KUBELESS_INSTALL_VOLUME)"},
					{Name: "FUNC_RUNTIME", Value: "python39"},
					{Name: "PYTHONUNBUFFERED", Value: "TRUE"}},
			},
		},
		"python312": {
			name:    "python312",
			runtime: serverlessv1alpha2.Python312,
			want: runtime.Config{
				Runtime:                 serverlessv1alpha2.Python312,
				DependencyFile:          "requirements.txt",
				FunctionFile:            "handler.py",
				DockerfileConfigMapName: "dockerfile-python312",
				RuntimeEnvs: []corev1.EnvVar{{Name: "PYTHONPATH", Value: "$(KUBELESS_INSTALL_VOLUME)/lib.python3.12/site-packages:$(KUBELESS_INSTALL_VOLUME)"},
					{Name: "FUNC_RUNTIME", Value: "python312"},
					{Name: "PYTHONUNBUFFERED", Value: "TRUE"}},
			},
		},
		"nodejs18": {
			name:    "nodejs18 config",
			runtime: serverlessv1alpha2.NodeJs18,
			want: runtime.Config{
				Runtime:                 serverlessv1alpha2.NodeJs18,
				DependencyFile:          "package.json",
				FunctionFile:            "handler.js",
				DockerfileConfigMapName: "dockerfile-nodejs18",
				RuntimeEnvs: []corev1.EnvVar{
					{Name: "FUNC_RUNTIME", Value: "nodejs18"}},
			},
		},
		"nodejs20": {
			name:    "nodejs20 config",
			runtime: serverlessv1alpha2.NodeJs20,
			want: runtime.Config{
				Runtime:                 serverlessv1alpha2.NodeJs20,
				DependencyFile:          "package.json",
				FunctionFile:            "handler.js",
				DockerfileConfigMapName: "dockerfile-nodejs20",
				RuntimeEnvs: []corev1.EnvVar{
					{Name: "FUNC_RUNTIME", Value: "nodejs20"}},
			},
		}} {
		t.Run(testName, func(t *testing.T) {
			//given
			g := gomega.NewWithT(t)

			// when
			config := runtime.GetRuntimeConfig(testData.runtime)

			// then
			// `RuntimeEnvs` may be in a different order, so I convert them to a map before comparing them
			configEnvMap := make(map[string]corev1.EnvVar)
			for _, ev := range config.RuntimeEnvs {
				configEnvMap[ev.Name] = ev
			}
			wantEnvMap := make(map[string]corev1.EnvVar)
			for _, ev := range testData.want.RuntimeEnvs {
				wantEnvMap[ev.Name] = ev
			}
			g.Expect(configEnvMap).To(gomega.BeEquivalentTo(wantEnvMap))

			// `RuntimeEnvs` were compared before, and now I want to compare the rest of `config`
			config.RuntimeEnvs = nil
			testData.want.RuntimeEnvs = nil
			g.Expect(config).To(gomega.BeEquivalentTo(testData.want))
		})
	}
}
