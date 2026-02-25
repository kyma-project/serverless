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

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

func HanaClient(restConfig *rest.Config, cfg internal.Config, logf *logrus.Entry) (executor.Step, error) {
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

	nodejs20Logger := logf.WithField(runtimeKey, "nodejs20")
	nodejs22Logger := logf.WithField(runtimeKey, "nodejs22")
	nodejs24Logger := logf.WithField(runtimeKey, "nodejs24")

	genericContainer := utils.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}

	nodejs20Fn := function.NewFunction("hana-nodejs20", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs20Logger))

	nodejs22Fn := function.NewFunction("hana-nodejs22", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs22Logger))

	nodejs24Fn := function.NewFunction("hana-nodejs24", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs24Logger))

	logf.Infof("Testing function in namespace: %s", cfg.Namespace)

	poll := utils.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		DataKey:            internal.TestDataKey,
	}

	return executor.NewSerialTestRunner(logf, "Runtime test",
		namespace.NewNamespaceStep(logf, fmt.Sprintf("Create %s namespace", genericContainer.Namespace), genericContainer.Namespace, coreCli),
		executor.NewParallelRunner(logf, "Fn tests",
			executor.NewSerialTestRunner(nodejs20Logger, "NodeJS20 test",
				function.CreateFunction(nodejs20Logger, nodejs20Fn, "Create NodeJS20 Function", runtimes.NodeJSFunctionUsingHanaClient(serverlessv1alpha2.NodeJs20)),
				assertion.NewHTTPCheck(nodejs20Logger, "Testing hana-client in nodejs20 function", nodejs20Fn.FunctionURL, poll, "OK"),
			),
			executor.NewSerialTestRunner(nodejs22Logger, "NodeJS22 test",
				function.CreateFunction(nodejs22Logger, nodejs22Fn, "Create NodeJS22 Function", runtimes.NodeJSFunctionUsingHanaClient(serverlessv1alpha2.NodeJs22)),
				assertion.NewHTTPCheck(nodejs22Logger, "Testing hana-client in nodejs22 function", nodejs22Fn.FunctionURL, poll, "OK"),
			),
			executor.NewSerialTestRunner(nodejs24Logger, "NodeJS24 test",
				function.CreateFunction(nodejs24Logger, nodejs24Fn, "Create NodeJS24 Function", runtimes.NodeJSFunctionUsingHanaClient(serverlessv1alpha2.NodeJs24)),
				assertion.NewHTTPCheck(nodejs24Logger, "Testing hana-client in nodejs24 function", nodejs24Fn.FunctionURL, poll, "OK"),
			),
		),
	), nil
}
