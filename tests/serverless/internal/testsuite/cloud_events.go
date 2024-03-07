package testsuite

import (
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kyma-project/serverless/tests/serverless/internal"
	"github.com/kyma-project/serverless/tests/serverless/internal/assertion"
	"github.com/kyma-project/serverless/tests/serverless/internal/executor"
	"github.com/kyma-project/serverless/tests/serverless/internal/resources/function"
	"github.com/kyma-project/serverless/tests/serverless/internal/resources/namespace"
	"github.com/kyma-project/serverless/tests/serverless/internal/resources/runtimes"
	"github.com/kyma-project/serverless/tests/serverless/internal/utils"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

func FunctionCloudEventsTest(restConfig *rest.Config, cfg internal.Config, logf *logrus.Entry) (executor.Step, error) {
	now := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%02dh%02dm%02ds", "test-cloud-events", now.Hour(), now.Minute(), now.Second())

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

	publisherProxyMock := function.NewFunction("eventing-publisher-proxy", "kyma-system", cfg.KubectlProxyEnabled, genericContainer.WithLogger(python312Logger))
	python39Fn := function.NewFunction("python39", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(python39Logger))
	python312Fn := function.NewFunction("python312", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(python312Logger))
	nodejs18Fn := function.NewFunction("nodejs18", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs18Logger))
	nodejs20Fn := function.NewFunction("nodejs20", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs20Logger))

	logf.Infof("Testing function in namespace: %s", cfg.Namespace)

	return executor.NewSerialTestRunner(logf, "Runtime test",
		namespace.NewNamespaceStep(logf, fmt.Sprintf("Create %s namespace", genericContainer.Namespace), genericContainer.Namespace, coreCli),
		function.CreateFunction(logf, publisherProxyMock, "Create publisher proxy mock", runtimes.PythonPublisherProxyMock()),
		executor.NewParallelRunner(logf, "Fn tests",
			executor.NewSerialTestRunner(python39Logger, "Python39 test",
				function.CreateFunction(python39Logger, python39Fn, "Create Python39 Function", runtimes.PythonCloudEvent(serverlessv1alpha2.Python39)),
				assertion.CloudEventReceiveCheck(python39Logger, "Python39 cloud event structured check", cloudevents.EncodingStructured, python39Fn.FunctionURL),
				assertion.CloudEventReceiveCheck(python39Logger, "Python39 cloud event binary check", cloudevents.EncodingBinary, python39Fn.FunctionURL),
				assertion.CloudEventSendCheck(python39Logger, "Python39 cloud event sent check", string(serverlessv1alpha2.Python39), python39Fn.FunctionURL, publisherProxyMock.FunctionURL),
			),
			executor.NewSerialTestRunner(python312Logger, "Python312 test",
				function.CreateFunction(python312Logger, python312Fn, "Create Python312 Function", runtimes.PythonCloudEvent(serverlessv1alpha2.Python312)),
				assertion.CloudEventReceiveCheck(python312Logger, "Python312 cloud event structured check", cloudevents.EncodingStructured, python312Fn.FunctionURL),
				assertion.CloudEventReceiveCheck(python312Logger, "Python312 cloud event binary check", cloudevents.EncodingBinary, python312Fn.FunctionURL),
				assertion.CloudEventSendCheck(python312Logger, "Python312 cloud event sent check", string(serverlessv1alpha2.Python312), python312Fn.FunctionURL, publisherProxyMock.FunctionURL),
			),
			executor.NewSerialTestRunner(nodejs18Logger, "NodeJS18 test",
				function.CreateFunction(nodejs18Logger, nodejs18Fn, "Create NodeJS18 Function", runtimes.NodeJSFunctionWithCloudEvent(serverlessv1alpha2.NodeJs18)),
				assertion.CloudEventReceiveCheck(nodejs18Logger, "NodeJS18 cloud event structured check", cloudevents.EncodingStructured, nodejs18Fn.FunctionURL),
				assertion.CloudEventReceiveCheck(nodejs18Logger, "NodeJS18 cloud event binary check", cloudevents.EncodingBinary, nodejs18Fn.FunctionURL),
				assertion.CloudEventSendCheck(nodejs18Logger, "NodeJS18 cloud event sent check", string(serverlessv1alpha2.NodeJs18), nodejs18Fn.FunctionURL, publisherProxyMock.FunctionURL),
			),
			executor.NewSerialTestRunner(nodejs20Logger, "NodeJS20 test",
				function.CreateFunction(nodejs20Logger, nodejs20Fn, "Create NodeJS20 Function", runtimes.NodeJSFunctionWithCloudEvent(serverlessv1alpha2.NodeJs20)),
				assertion.CloudEventReceiveCheck(nodejs20Logger, "NodeJS20 cloud event structured check", cloudevents.EncodingStructured, nodejs20Fn.FunctionURL),
				assertion.CloudEventReceiveCheck(nodejs20Logger, "NodeJS20 cloud event binary check", cloudevents.EncodingBinary, nodejs20Fn.FunctionURL),
				assertion.CloudEventSendCheck(nodejs20Logger, "NodeJS20 cloud event sent check", string(serverlessv1alpha2.NodeJs20), nodejs20Fn.FunctionURL, publisherProxyMock.FunctionURL),
			),
		),
	), nil
}
