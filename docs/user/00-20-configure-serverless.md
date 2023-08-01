# Overview

Serverless module has its own operator (`serverless-operator`). It watches `Serverless` Custom Resource and re-configure (reconcile) serverless workloads.

Serverless CR becomes the API to configure serverless module. Use it to do the following:
 - enable/disable internal docker registry
 - configure external docker registry 
 - override endpoint for traces collected by serverless functions
 - override endpoint for eventing

Default configuration of serverless module looks as follows:

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: Serverless
  name: serverless-sample
spec:
  dockerRegistry:
    enableInternal: true
```

## Configure Docker Registry

By default, Serverless uses PersistentVolume (PV) as the internal registry to store Docker images for Functions. The default storage size of a single volume is 20 GB. This internal registry is suitable for local development.

If you use Serverless for production purposes, it is recommended that you use an external registry, such as Docker Hub, Google Container Registry (GCR), or Azure Container Registry (ACR).

If you want serverless to use external docker registry, create a secret in `kyma-system` Namespace. Such a Secret must contain required data (**username**, **password**, **serverAddress**, and **registryAddress**):

```bash
kubectl create secret -n kyma-system generic my-registry-config --from-literal=username={your-docker-reg-username} --from-literal=password={your-docker-reg-password} --from-literal=serverAddress={your-docker-reg-server-url}  --from-literal=registryAddress={your-docker-reg-registry-url}
```

>**TIP:** In case of DockerHub, usually the Docker registry address is the same as the account name.

Example:

```bash
kubectl create secret -n kyma-system generic my-registry-config --from-literal=username=kyma-rocks --from-literal=password=admin123 --from-literal=serverAddress=https://index.docker.io/v1/  --from-literal=registryAddress=kyma-rocks
```
Than, reference the secret in the serverless CR

```yaml
spec:
  dockerRegistry:
    secretName: my-registry-config 
```
URL of currently used docker registry is visible in the serverless CR status.


## Configure Trace Endpoint

By default, serverless operator detects if there is an available trace endpoint available by inspecting `TracePieline` CR. If it is available, the detected trace endpoint is used as (trace collector URL) in functions.
If no tracePipeline CR is detected functions are configured with no trace collector endpoint.
You can configure custom trace endpoint, so that function traces are sent over to any tracing backend of your choice.
Currently used trace endpoint is visible in the serverless CR status.

```yaml
spec:
  tracing:
    endpoint: http://tracing-jaeger-collector.kyma-system.svc.cluster.local:2342/v1/metrics 
```

## Configure Eventing endpoint

You can configure custom eventing endpoint, so that, when you use SDK for sending events from your functions, it is used to publish events to.
Currently used trace endpoint is visible in the serverless CR status.
By default `http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish` is used.

```yaml
spec:
  eventing:
    endpoint: http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish
```
