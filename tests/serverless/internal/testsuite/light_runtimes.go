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

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

const runtimeKey = "runtime"

func SimpleFunctionTest(restConfig *rest.Config, cfg internal.Config, logf *logrus.Entry) (executor.Step, error) {
	now := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%02dh%02dm%02ds", "test-serverless-simple", now.Hour(), now.Minute(), now.Second())

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating dynamic client")
	}

	coreCli, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s CoreV1Client")
	}

	python39Logger := logf.WithField(runtimeKey, "python39")
	python312Logger := logf.WithField(runtimeKey, "python312")
	nodejs18Logger := logf.WithField(runtimeKey, "nodejs18")
	nodejs20Logger := logf.WithField(runtimeKey, "nodejs20")

	genericContainer := utils.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}

	python39Fn := function.NewFunction("python39", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(python39Logger))

	python312Fn := function.NewFunction("python312", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(python312Logger))

	nodejs18Fn := function.NewFunction("nodejs18", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs18Logger))

	nodejs20Fn := function.NewFunction("nodejs20", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs20Logger))

	cm := configmap.NewConfigMap("test-serverless-configmap", genericContainer.WithLogger(nodejs20Logger))
	cmEnvKey := "CM_ENV_KEY"
	cmEnvValue := "Value taken as env from ConfigMap"
	cmData := map[string]string{
		cmEnvKey: cmEnvValue,
	}
	sec := secret.NewSecret("test-serverless-secret", genericContainer.WithLogger(nodejs20Logger))
	secEnvKey := "SECRET_ENV_KEY"
	secEnvValue := "Value taken as env from Secret"
	secretData := map[string]string{
		secEnvKey: secEnvValue,
	}

	pkgCfgSecret := secret.NewSecret(cfg.PackageRegistryConfigSecretName, genericContainer)
	pkgCfgSecretData := map[string]string{
		".npmrc":   fmt.Sprintf("@kyma:registry=%s\nalways-auth=true", cfg.PackageRegistryConfigURLNode),
		"pip.conf": fmt.Sprintf("[global]\nextra-index-url = %s", cfg.PackageRegistryConfigURLPython),
	}

	logf.Infof("Testing function in namespace: %s", cfg.Namespace)

	poll := utils.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		DataKey:            internal.TestDataKey,
	}
	return executor.NewSerialTestRunner(logf, "Runtime test",
		namespace.NewNamespaceStep(logf, fmt.Sprintf("Create %s namespace", genericContainer.Namespace), genericContainer.Namespace, coreCli),
		secret.CreateSecret(logf, pkgCfgSecret, "Create package configuration secret", pkgCfgSecretData),
		executor.NewParallelRunner(logf, "Fn tests",
			executor.NewSerialTestRunner(python39Logger, "Python39 test",
				function.CreateFunction(python39Logger, python39Fn, "Create Python39 Function", runtimes.BasicPythonFunction("Hello From python", serverlessv1alpha2.Python39)),
				assertion.NewHTTPCheck(python39Logger, "Python39 pre update simple check through service", python39Fn.FunctionURL, poll, "Hello From python"),
				function.UpdateFunction(python39Logger, python39Fn, "Update Python39 Function", runtimes.BasicPythonFunctionWithCustomDependency("Hello From updated python", serverlessv1alpha2.Python39)),
				assertion.NewHTTPCheck(python39Logger, "Python39 post update simple check through service", python39Fn.FunctionURL, poll, "Hello From updated python"),
			),
			executor.NewSerialTestRunner(python312Logger, "Python312 test",
				function.CreateFunction(python312Logger, python312Fn, "Create Python312 Function", runtimes.BasicPythonFunction("Hello From python", serverlessv1alpha2.Python312)),
				assertion.NewHTTPCheck(python312Logger, "Python312 pre update simple check through service", python312Fn.FunctionURL, poll, "Hello From python"),
				function.UpdateFunction(python312Logger, python312Fn, "Update Python312 Function", runtimes.BasicPythonFunctionWithCustomDependency("Hello From updated python", serverlessv1alpha2.Python312)),
				assertion.NewHTTPCheck(python312Logger, "Python312 post update simple check through service", python312Fn.FunctionURL, poll, "Hello From updated python"),
			),
			executor.NewSerialTestRunner(nodejs18Logger, "NodeJS18 test",
				configmap.CreateConfigMap(nodejs18Logger, cm, "Create Test ConfigMap", cmData),
				secret.CreateSecret(nodejs18Logger, sec, "Create Test Secret", secretData),
				function.CreateFunction(nodejs18Logger, nodejs18Fn, "Create NodeJS18 Function", runtimes.NodeJSFunctionWithEnvFromConfigMapAndSecret(cm.Name(), cmEnvKey, sec.Name(), secEnvKey, serverlessv1alpha2.NodeJs18)),
				assertion.NewHTTPCheck(nodejs18Logger, "NodeJS18 pre update simple check through service", nodejs18Fn.FunctionURL, poll, fmt.Sprintf("%s-%s", cmEnvValue, secEnvValue)),
				function.UpdateFunction(nodejs18Logger, nodejs18Fn, "Update NodeJS18 Function", runtimes.BasicNodeJSFunctionWithCustomDependency("Hello from updated nodejs18", serverlessv1alpha2.NodeJs18)),
				assertion.NewHTTPCheck(nodejs18Logger, "NodeJS18 post update simple check through service", nodejs18Fn.FunctionURL, poll, "Hello from updated nodejs18"),
			),
			executor.NewSerialTestRunner(nodejs20Logger, "NodeJS20 test",
				function.CreateFunction(nodejs20Logger, nodejs20Fn, "Create NodeJS20 Function", runtimes.NodeJSFunctionWithEnvFromConfigMapAndSecret(cm.Name(), cmEnvKey, sec.Name(), secEnvKey, serverlessv1alpha2.NodeJs20)),
				assertion.NewHTTPCheck(nodejs20Logger, "NodeJS20 pre update simple check through service", nodejs20Fn.FunctionURL, poll, fmt.Sprintf("%s-%s", cmEnvValue, secEnvValue)),
				function.UpdateFunction(nodejs20Logger, nodejs20Fn, "Update NodeJS20 Function", runtimes.BasicNodeJSFunctionWithCustomDependency("Hello from updated nodejs20", serverlessv1alpha2.NodeJs20)),
				assertion.NewHTTPCheck(nodejs20Logger, "NodeJS20 post update simple check through service", nodejs20Fn.FunctionURL, poll, "Hello from updated nodejs20"),
			),
		),
	), nil
}
