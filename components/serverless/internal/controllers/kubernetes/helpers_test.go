package kubernetes

import (
	"github.com/onsi/gomega"
	"github.com/vrischmann/envconfig"
	corev1 "k8s.io/api/core/v1"
)

func setUpControllerConfig(g *gomega.GomegaWithT) Config {
	var testCfg Config
	err := envconfig.InitWithPrefix(&testCfg, "TEST")
	g.Expect(err).To(gomega.BeNil())
	return testCfg
}

func compareConfigMaps(g *gomega.WithT, actual, expected *corev1.ConfigMap) {
	g.Expect(actual.GetLabels()).To(gomega.Equal(expected.GetLabels()))
	g.Expect(actual.GetAnnotations()).To(gomega.Equal(expected.GetAnnotations()))
	g.Expect(actual.Data).To(gomega.Equal(expected.Data))
	g.Expect(actual.BinaryData).To(gomega.Equal(expected.BinaryData))
}

func compareSecrets(g *gomega.WithT, actual, expected *corev1.Secret) {
	g.Expect(actual.GetLabels()).To(gomega.Equal(expected.GetLabels()))
	g.Expect(actual.GetAnnotations()).To(gomega.Equal(expected.GetAnnotations()))
	g.Expect(actual.Data).To(gomega.Equal(expected.Data))
}

func compareServiceAccounts(g *gomega.WithT, actual, expected *corev1.ServiceAccount) {
	g.Expect(actual.GetLabels()).To(gomega.Equal(expected.GetLabels()))
	g.Expect(actual.GetAnnotations()).To(gomega.Equal(expected.GetAnnotations()))
	g.Expect(actual.Secrets).To(gomega.Equal(expected.Secrets))
	g.Expect(actual.ImagePullSecrets).To(gomega.Equal(expected.ImagePullSecrets))
	g.Expect(actual.AutomountServiceAccountToken).To(gomega.Equal(expected.AutomountServiceAccountToken))
}
