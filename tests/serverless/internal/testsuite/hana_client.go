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

func HanaClientTest(restConfig *rest.Config, cfg internal.Config, logf *logrus.Entry) (executor.Step, error) {
	now := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%02dh%02dm%02ds", "test-hana-client", now.Hour(), now.Minute(), now.Second())

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating dynamic client")
	}

	coreCli, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s CoreV1Client")
	}

	nodejs18Logger := logf.WithField(runtimeKey, "nodejs18")
	nodejs20Logger := logf.WithField(runtimeKey, "nodejs20")

	genericContainer := utils.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}

	nodejs18Fn := function.NewFunction("hana-nodejs18", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs18Logger))

	nodejs20Fn := function.NewFunction("hana-nodejs20", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs20Logger))

	logf.Infof("Testing function in namespace: %s", cfg.Namespace)

	poll := utils.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		DataKey:            internal.TestDataKey,
	}

	return executor.NewSerialTestRunner(logf, "Runtime test",
		namespace.NewNamespaceStep(logf, fmt.Sprintf("Create %s namespace", genericContainer.Namespace), genericContainer.Namespace, coreCli),
		executor.NewParallelRunner(logf, "Fn tests",
			executor.NewSerialTestRunner(nodejs18Logger, "NodeJS18 test",
				function.CreateFunction(nodejs18Logger, nodejs18Fn, "Create NodeJS18 Function", runtimes.NodeJSFunctionUsingHanaClient(serverlessv1alpha2.NodeJs18)),
				assertion.NewHTTPCheck(nodejs18Logger, "Testing hana-client in nodejs18 function", nodejs18Fn.FunctionURL, poll, "{\"code\":10,\"sqlState\":\"28000\"}"),
			),
			executor.NewSerialTestRunner(nodejs20Logger, "NodeJS20 test",
				function.CreateFunction(nodejs20Logger, nodejs20Fn, "Create NodeJS20 Function", runtimes.NodeJSFunctionUsingHanaClient(serverlessv1alpha2.NodeJs20)),
				assertion.NewHTTPCheck(nodejs18Logger, "Testing hana-client in nodejs20 function", nodejs20Fn.FunctionURL, poll, "{\"code\":10,\"sqlState\":\"28000\"}"),
			),
		),
	), nil
}
