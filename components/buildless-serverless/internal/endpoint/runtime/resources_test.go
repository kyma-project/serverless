package runtime

import (
	"encoding/base64"
	"testing"

	"github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/config"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildResources(t *testing.T) {
	t.Run("build resources for function", func(t *testing.T) {
		files, err := BuildResources(&config.FunctionConfig{}, &v1alpha2.Function{
			Spec: v1alpha2.FunctionSpec{
				Runtime: "nodejs22",
				Source: v1alpha2.Source{
					Inline: &v1alpha2.InlineSource{
						Source:       "console.log('Hello World')",
						Dependencies: "{}",
					},
				},
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-function",
				Namespace: "test-namespace",
			},
		})

		require.NoError(t, err)
		require.Len(t, files, 2)
		require.Equal(t, "resources/service.yaml", files[0].Name)
		requireEqualBase64Objects(t, fixTestService(), files[0].Data)
		require.Equal(t, "resources/deployment.yaml", files[1].Name)
		requireEqualBase64Objects(t, fixDeployment(), files[1].Data)
	})

	t.Run("error on git source", func(t *testing.T) {
		files, err := BuildResources(&config.FunctionConfig{}, &v1alpha2.Function{
			Spec: v1alpha2.FunctionSpec{
				Runtime: "nodejs22",
				Source: v1alpha2.Source{
					GitRepository: &v1alpha2.GitRepositorySource{},
				},
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-function",
				Namespace: "test-namespace",
			},
		})

		require.ErrorContains(t, err, "ejecting functions with git source is not supported")
		require.Nil(t, files)
	})
}

func fixTestService() string {
	return `apiVersion: v1
kind: Service
metadata:
  creationTimestamp: null
  name: test-function-ejected
  namespace: test-namespace
spec:
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: 8080
  selector:
    app.kubernetes.io/instance: test-function-ejected
    serverless.kyma-project.io/resource: deployment
status:
  loadBalancer: {}
`
}

func fixDeployment() string {
	return `apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  name: test-function-ejected
  namespace: test-namespace
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/instance: test-function-ejected
      serverless.kyma-project.io/resource: deployment
  strategy: {}
  template:
    metadata:
      annotations:
        proxy.istio.io/config: '{ "holdApplicationUntilProxyStarts": true }'
        sidecar.istio.io/nativeSidecar: "true"
      creationTimestamp: null
      labels:
        app.kubernetes.io/instance: test-function-ejected
        app.kubernetes.io/name: test-function
        serverless.kyma-project.io/resource: deployment
    spec:
      containers:
      - env:
        - name: FUNC_NAME
          value: test-function
        - name: FUNC_RUNTIME
          value: nodejs22
        - name: SERVICE_NAMESPACE
          value: test-namespace
        - name: TRACE_COLLECTOR_ENDPOINT
        - name: PUBLISHER_PROXY_ADDRESS
        image: image:tag
        livenessProbe:
          failureThreshold: 3
          httpGet:
            path: /healthz
            port: 8080
          periodSeconds: 5
          timeoutSeconds: 4
        name: function
        ports:
        - containerPort: 8080
          protocol: TCP
        readinessProbe:
          failureThreshold: 1
          httpGet:
            path: /healthz
            port: 8080
          periodSeconds: 5
          timeoutSeconds: 2
        resources: {}
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          privileged: false
          procMount: Default
          readOnlyRootFilesystem: false
          runAsNonRoot: true
        startupProbe:
          failureThreshold: 30
          httpGet:
            path: /healthz
            port: 8080
          periodSeconds: 5
          successThreshold: 1
        volumeMounts:
        - mountPath: /usr/src/app/function
          name: sources
        - mountPath: /tmp
          name: tmp
        - mountPath: /usr/src/app/function/package-registry-config/.npmrc
          name: package-registry-config
          subPath: .npmrc
        workingDir: /usr/src/app/function
      securityContext:
        runAsGroup: 1000
        runAsUser: 1000
        seccompProfile:
          type: RuntimeDefault
      volumes:
      - emptyDir: {}
        name: sources
      - name: package-registry-config
        secret:
          optional: true
      - emptyDir: {}
        name: tmp
status: {}
`
}

func requireEqualBase64Objects(t *testing.T, expectedObj, actual string) {
	actualBytes, err := base64.StdEncoding.DecodeString(actual)
	require.NoError(t, err)

	require.Equal(t,
		expectedObj,
		string(actualBytes),
	)
}
