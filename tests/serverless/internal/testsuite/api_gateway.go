package testsuite

import (
	"fmt"
	"time"

	"github.com/kyma-project/serverless/tests/serverless/internal"
	"github.com/kyma-project/serverless/tests/serverless/internal/assertion"
	"github.com/kyma-project/serverless/tests/serverless/internal/executor"
	"github.com/kyma-project/serverless/tests/serverless/internal/resources/function"
	"github.com/kyma-project/serverless/tests/serverless/internal/resources/namespace"
	"github.com/kyma-project/serverless/tests/serverless/internal/resources/runtimes"
	"github.com/kyma-project/serverless/tests/serverless/internal/utils"
	"github.com/pkg/errors"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

const (
	nodejs18  = "nodejs18"
	nodejs20  = "nodejs20"
	python39  = "python39"
	python312 = "python312"
)

func FunctionAPIGatewayTest(restConfig *rest.Config, cfg internal.Config, logf *logrus.Entry) (executor.Step, error) {
	now := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%02dh%02dm%02ds", "test-api-gateway", now.Hour(), now.Minute(), now.Second())

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating dynamic client")
	}

	coreCli, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s CoreV1Client")
	}

	python312Logger := logf.WithField(runtimeKey, "python312")
	python39Logger := logf.WithField(runtimeKey, "python39")
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

	logf.Infof("Testing function in namespace: %s", cfg.Namespace)

	return executor.NewSerialTestRunner(logf, "Runtime test",
		namespace.NewNamespaceStep(logf, fmt.Sprintf("Create %s namespace", genericContainer.Namespace), genericContainer.Namespace, coreCli),
		executor.NewParallelRunner(logf, "Fn tests",
			executor.NewSerialTestRunner(python39Logger, "Python39 test",
				function.CreateFunction(python39Logger, python39Fn, "Create Python39 Function", runtimes.BasicPythonFunction("Hello from python39", serverlessv1alpha2.Python39)),
				assertion.APIGatewayFunctionCheck("python39", python39Fn, coreCli, genericContainer.Namespace, python39),
			),
			executor.NewSerialTestRunner(python312Logger, "Python312 test",
				function.CreateFunction(python312Logger, python312Fn, "Create Python312 Function", runtimes.BasicPythonFunction("Hello from python312", serverlessv1alpha2.Python312)),
				assertion.APIGatewayFunctionCheck("python312", python312Fn, coreCli, genericContainer.Namespace, python312),
			),
			executor.NewSerialTestRunner(nodejs18Logger, "NodeJS18 test",
				function.CreateFunction(nodejs18Logger, nodejs18Fn, "Create NodeJS18 Function", runtimes.BasicNodeJSFunction("Hello from nodejs18", serverlessv1alpha2.NodeJs18)),
				assertion.APIGatewayFunctionCheck("nodejs18", nodejs18Fn, coreCli, genericContainer.Namespace, nodejs18),
			),
			executor.NewSerialTestRunner(nodejs20Logger, "NodeJS20 test",
				function.CreateFunction(nodejs20Logger, nodejs20Fn, "Create NodeJS20 Function", runtimes.BasicNodeJSFunction("Hello from nodejs20", serverlessv1alpha2.NodeJs20)),
				assertion.APIGatewayFunctionCheck("nodejs20", nodejs20Fn, coreCli, genericContainer.Namespace, nodejs20),
			),
		),
	), nil
}
