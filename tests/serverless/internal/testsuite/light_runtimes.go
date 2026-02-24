package testsuite

import (
	"fmt"
	"time"

	"github.com/kyma-project/serverless/tests/serverless/internal"
	"github.com/kyma-project/serverless/tests/serverless/internal/assertion"
	"github.com/kyma-project/serverless/tests/serverless/internal/executor"
	"github.com/kyma-project/serverless/tests/serverless/internal/resources/configmap"
	"github.com/kyma-project/serverless/tests/serverless/internal/resources/function"
	"github.com/kyma-project/serverless/tests/serverless/internal/resources/namespace"
	"github.com/kyma-project/serverless/tests/serverless/internal/resources/runtimes"
	"github.com/kyma-project/serverless/tests/serverless/internal/resources/secret"
	"github.com/kyma-project/serverless/tests/serverless/internal/utils"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

const (
	runtimeKey = "runtime"
	cmEnvKey   = "CM_ENV_KEY"
	cmEnvValue = "Value taken as env from ConfigMap"

	secEnvKey   = "SECRET_ENV_KEY"
	secEnvValue = "Value taken as env from Secret"
)

var (
	secretData = map[string]string{
		secEnvKey: secEnvValue,
	}
	cmData = map[string]string{
		cmEnvKey: cmEnvValue,
	}
)

func SimpleFunctionTest(restConfig *rest.Config, cfg internal.Config, logf *logrus.Entry) (executor.Step, error) {
	now := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%02dh%02dm%02ds", "test-serverless-simple", now.Hour(), now.Minute(), now.Second())

	coreCli, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s CoreV1Client")
	}

	genericContainer, err := newGenericContainer(logf, restConfig, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while creating generic container")
	}

	logf.Infof("Testing function in namespace: %s", cfg.Namespace)

	poll := utils.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		DataKey:            internal.TestDataKey,
	}
	return executor.NewSerialTestRunner(logf, "Runtime test",
		namespace.NewNamespaceStep(logf, fmt.Sprintf("Create %s namespace", genericContainer.Namespace), genericContainer.Namespace, coreCli),
		newRegistryConfigSecretStep(logf, genericContainer, cfg),
		executor.NewParallelRunner(logf, "Fn tests",
			newPython312TestRunner(logf, poll, genericContainer, cfg.KubectlProxyEnabled),
			newNodejs20TestRunner(logf, poll, genericContainer, cfg.KubectlProxyEnabled),
			newNodejs22TestRunner(logf, poll, genericContainer, cfg.KubectlProxyEnabled),
			newNodejs24TestRunner(logf, poll, genericContainer, cfg.KubectlProxyEnabled),
		),
	), nil
}

func FIPSSimpleFunctionTest(restConfig *rest.Config, cfg internal.Config, logf *logrus.Entry) (executor.Step, error) {
	now := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%02dh%02dm%02ds", "test-serverless-simple", now.Hour(), now.Minute(), now.Second())

	coreCli, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s CoreV1Client")
	}

	genericContainer, err := newGenericContainer(logf, restConfig, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while creating generic container")
	}

	logf.Infof("Testing function in namespace: %s", cfg.Namespace)

	poll := utils.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		DataKey:            internal.TestDataKey,
	}
	return executor.NewSerialTestRunner(logf, "Runtime test",
		namespace.NewNamespaceStep(logf, fmt.Sprintf("Create %s namespace", genericContainer.Namespace), genericContainer.Namespace, coreCli),
		newRegistryConfigSecretStep(logf, genericContainer, cfg),
		executor.NewParallelRunner(logf, "Fn tests",
			newNodejs22TestRunner(logf, poll, genericContainer, cfg.KubectlProxyEnabled),
			newNodejs24TestRunner(logf, poll, genericContainer, cfg.KubectlProxyEnabled),
		),
	), nil
}

func newGenericContainer(logf *logrus.Entry, restConfig *rest.Config, cfg internal.Config) (utils.Container, error) {
	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return utils.Container{}, err
	}

	return utils.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}, nil
}

func newRegistryConfigSecretStep(logf *logrus.Entry, genericContainer utils.Container, cfg internal.Config) executor.Step {
	pkgCfgSecret := secret.NewSecret(cfg.PackageRegistryConfigSecretName, genericContainer)
	pkgCfgSecretData := map[string]string{
		".npmrc":   fmt.Sprintf("@kyma:registry=%s\nalways-auth=true", cfg.PackageRegistryConfigURLNode),
		"pip.conf": fmt.Sprintf("[global]\nextra-index-url = %s", cfg.PackageRegistryConfigURLPython),
	}

	return secret.CreateSecret(logf, pkgCfgSecret, "Create package configuration secret", pkgCfgSecretData)
}

func newNodejs20TestRunner(logf *logrus.Entry, poll utils.Poller, genericContainer utils.Container, kubectlProxyEnabled bool) *executor.SerialRunner {
	nodejs20Logger := logf.WithField(runtimeKey, "nodejs20")
	nodejs20Fn := function.NewFunction("nodejs20", genericContainer.Namespace, kubectlProxyEnabled, genericContainer.WithLogger(nodejs20Logger))
	cmNodeJS20 := configmap.NewConfigMap("test-serverless-configmap-nodejs20", genericContainer.WithLogger(nodejs20Logger))
	secNodeJS20 := secret.NewSecret("test-serverless-secret-nodejs20", genericContainer.WithLogger(nodejs20Logger))

	return executor.NewSerialTestRunner(nodejs20Logger, "NodeJS20 test",
		configmap.CreateConfigMap(nodejs20Logger, cmNodeJS20, "Create Test ConfigMap", cmData),
		secret.CreateSecret(nodejs20Logger, secNodeJS20, "Create Test Secret", secretData),
		function.CreateFunction(nodejs20Logger, nodejs20Fn, "Create NodeJS20 Function", runtimes.NodeJSFunctionWithEnvFromConfigMapAndSecret(cmNodeJS20.Name(), cmEnvKey, secNodeJS20.Name(), secEnvKey, serverlessv1alpha2.NodeJs20)),
		assertion.NewHTTPCheck(nodejs20Logger, "NodeJS20 pre update simple check through service", nodejs20Fn.FunctionURL, poll, fmt.Sprintf("%s-%s", cmEnvValue, secEnvValue)),
		function.UpdateFunction(nodejs20Logger, nodejs20Fn, "Update NodeJS20 Function", runtimes.BasicNodeJSFunctionWithCustomDependency("Hello from updated nodejs20", serverlessv1alpha2.NodeJs20)),
		assertion.NewHTTPCheck(nodejs20Logger, "NodeJS20 post update simple check through service", nodejs20Fn.FunctionURL, poll, "Hello from updated nodejs20"),
	)
}

func newNodejs22TestRunner(logf *logrus.Entry, poll utils.Poller, genericContainer utils.Container, kubectlProxyEnabled bool) *executor.SerialRunner {
	nodejs22Logger := logf.WithField(runtimeKey, "nodejs22")
	nodejs22Fn := function.NewFunction("nodejs22", genericContainer.Namespace, kubectlProxyEnabled, genericContainer.WithLogger(nodejs22Logger))
	cmNodeJS22 := configmap.NewConfigMap("test-serverless-configmap-nodejs22", genericContainer.WithLogger(nodejs22Logger))
	secNodeJS22 := secret.NewSecret("test-serverless-secret-nodejs22", genericContainer.WithLogger(nodejs22Logger))

	return executor.NewSerialTestRunner(nodejs22Logger, "NodeJS22 test",
		configmap.CreateConfigMap(nodejs22Logger, cmNodeJS22, "Create Test ConfigMap", cmData),
		secret.CreateSecret(nodejs22Logger, secNodeJS22, "Create Test Secret", secretData),
		function.CreateFunction(nodejs22Logger, nodejs22Fn, "Create NodeJS22 Function", runtimes.NodeJSFunctionWithEnvFromConfigMapAndSecret(cmNodeJS22.Name(), cmEnvKey, secNodeJS22.Name(), secEnvKey, serverlessv1alpha2.NodeJs22)),
		assertion.NewHTTPCheck(nodejs22Logger, "NodeJS22 pre update simple check through service", nodejs22Fn.FunctionURL, poll, fmt.Sprintf("%s-%s", cmEnvValue, secEnvValue)),
		function.UpdateFunction(nodejs22Logger, nodejs22Fn, "Update NodeJS22 Function", runtimes.BasicNodeJSFunctionWithCustomDependency("Hello from updated nodejs22", serverlessv1alpha2.NodeJs22)),
		assertion.NewHTTPCheck(nodejs22Logger, "NodeJS22 post update simple check through service", nodejs22Fn.FunctionURL, poll, "Hello from updated nodejs22"),
	)
}

func newNodejs24TestRunner(logf *logrus.Entry, poll utils.Poller, genericContainer utils.Container, kubectlProxyEnabled bool) *executor.SerialRunner {
	nodejs24Logger := logf.WithField(runtimeKey, "nodejs24")
	nodejs24Fn := function.NewFunction("nodejs24", genericContainer.Namespace, kubectlProxyEnabled, genericContainer.WithLogger(nodejs24Logger))
	cmNodeJS24 := configmap.NewConfigMap("test-serverless-configmap-nodejs24", genericContainer.WithLogger(nodejs24Logger))
	secNodeJS24 := secret.NewSecret("test-serverless-secret-nodejs24", genericContainer.WithLogger(nodejs24Logger))

	return executor.NewSerialTestRunner(nodejs24Logger, "NodeJS24 test",
		configmap.CreateConfigMap(nodejs24Logger, cmNodeJS24, "Create Test ConfigMap", cmData),
		secret.CreateSecret(nodejs24Logger, secNodeJS24, "Create Test Secret", secretData),
		function.CreateFunction(nodejs24Logger, nodejs24Fn, "Create NodeJS24 Function", runtimes.NodeJSFunctionWithEnvFromConfigMapAndSecret(cmNodeJS24.Name(), cmEnvKey, secNodeJS24.Name(), secEnvKey, serverlessv1alpha2.NodeJs24)),
		assertion.NewHTTPCheck(nodejs24Logger, "NodeJS24 pre update simple check through service", nodejs24Fn.FunctionURL, poll, fmt.Sprintf("%s-%s", cmEnvValue, secEnvValue)),
		function.UpdateFunction(nodejs24Logger, nodejs24Fn, "Update NodeJS24 Function", runtimes.BasicNodeJSFunctionWithCustomDependency("Hello from updated nodejs24", serverlessv1alpha2.NodeJs24)),
		assertion.NewHTTPCheck(nodejs24Logger, "NodeJS24 post update simple check through service", nodejs24Fn.FunctionURL, poll, "Hello from updated nodejs24"),
	)
}

func newPython312TestRunner(logf *logrus.Entry, poll utils.Poller, genericContainer utils.Container, kubectlProxyEnabled bool) *executor.SerialRunner {
	python312Logger := logf.WithField(runtimeKey, "python312")
	python312Fn := function.NewFunction("python312", genericContainer.Namespace, kubectlProxyEnabled, genericContainer.WithLogger(python312Logger))

	return executor.NewSerialTestRunner(python312Logger, "Python312 test",
		function.CreateFunction(python312Logger, python312Fn, "Create Python312 Function", runtimes.BasicPythonFunction("Hello From python", serverlessv1alpha2.Python312)),
		assertion.NewHTTPCheck(python312Logger, "Python312 pre update simple check through service", python312Fn.FunctionURL, poll, "Hello From python"),
		function.UpdateFunction(python312Logger, python312Fn, "Update Python312 Function", runtimes.BasicPythonFunctionWithCustomDependency("Hello From updated python", serverlessv1alpha2.Python312)),
		assertion.NewHTTPCheck(python312Logger, "Python312 post update simple check through service", python312Fn.FunctionURL, poll, "Hello From updated python"),
	)
}
