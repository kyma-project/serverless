# Overview

The Serverless module has its own operator (Serverless operator). It watches the Serverless custom resource (CR) and reconfigures (reconciles) the Serverless workloads.

The Serverless CR becomes an API to configure the Serverless module. You can use it to:
 - enable or disable the internal Docker registry
 - configure the external Docker registry 
 - override endpoint for traces collected by the Serverless Functions
 - override endpoint for eventing

The default configuration of the Serverless Module is following:

   ```yaml
   apiVersion: operator.kyma-project.io/v1alpha1
   kind: Serverless
     name: serverless-sample
   spec:
     dockerRegistry:
       enableInternal: true
   ```

## Configure Docker registry

By default, Serverless uses PersistentVolume (PV) as the internal registry to store Docker images for Functions. The default storage size of a single volume is 20 GB. This internal registry is suitable for local development.

If you use Serverless for production purposes, it is recommended that you use an external registry, such as Docker Hub, Google Container Registry (GCR), or Azure Container Registry (ACR).

Follow these steps to use the external Docker registry in Serverless: 

1. Create a Secret in the `kyma-system` Namespace with the required data (`username`, `password`, `serverAddress`, and `registryAddress`):

   ```bash
   kubectl create secret -n kyma-system generic my-registry-config --from-literal=username={your-docker-reg-username} --from-literal=password={your-docker-reg-password} --from-literal=serverAddress={your-docker-reg-server-url}  --from-literal=registryAddress={your-docker-reg-registry-url}
   ```

>**TIP:** In case of DockerHub, usually the Docker registry address is the same as the account name.

Example:

   ```bash
   kubectl create secret -n kyma-system generic my-registry-config --from-literal=username=kyma-rocks --from-literal=password=admin123 --from-literal=serverAddress=https://index.docker.io/v1/  --from-literal=registryAddress=kyma-rocks
   ```
2. Reference the Secret in the Serverless CR

   ```yaml
   spec:
     dockerRegistry:
       secretName: my-registry-config 
   ```
The URL of the currently used Docker registry is visible in the Serverless CR status.


## Configure trace endpoint

By default, the Serverless operator checks if there is a trace endpoint available. If available, the detected trace endpoint is used as the trace collector URL in Functions.
If no trace endpoint is detected, Functions are configured with no trace collector endpoint.
You can configure a custom trace endpoint so that Function traces are sent to any tracing backend you choose.
The currently used trace endpoint is visible in the Serverless CR status.

   ```yaml
   spec:
     tracing:
       endpoint: http://tracing-jaeger-collector.kyma-system.svc.cluster.local:2342/v1/metrics 
   ```

## Configure Eventing endpoint

You can configure a custom eventing endpoint, so when you use SDK for sending events from your Functions, it is used to publish events.
The currently used trace endpoint is visible in the Serverless CR status.
By default `http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish` is used.

   ```yaml
   spec:
     eventing:
       endpoint: http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish
   ```
