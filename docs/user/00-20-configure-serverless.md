# Serverless configuration

## Overview

The Serverless module has its own operator (Serverless operator). It watches the Serverless custom resource (CR) and reconfigures (reconciles) the Serverless workloads.

The Serverless CR becomes an API to configure the Serverless module. You can use it to:
 - enable or disable the internal Docker registry
 - configure the external Docker registry 
 - override endpoint for traces collected by the Serverless Functions
 - override endpoint for eventing
 - override target CPU utilization percentage
 - override the Function requeue duration
 - override the Function build executor arguments
 - override function build max simultaneous jobs
 - override healthz liveness timeout
 - override function request body limit 
 - override function timeout
 - override default build job preset
 - override default runtime pod preset

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

## Configure target CPU utilization percentage

You can set a custom target threshold for CPU utilization. The default value is set to 50%.

```yaml
   spec:
      targetCPUUtilizationPercentage: 50
```

## Configure function requeue duration

By default, the function associated with default configuration will be requeued every 5 minutes.  

```yaml
   spec:
      functionRequeueDuration: 5m
```

## Configure function build executor arguments

Using this label you can choose the arguments passed to function build executor, for example: 
- `--insecure` - executor should operate in an insecure mode
- `--skip-tls-verify` - executor should skip TLS certificate verification
- `--skip-unused-stages` - executor should skip any stages aren't used for the current execution
- `--log-format=text` - executor should use logs in given format
- `--cache=true` - enables caching for the executor

```yaml
   spec:
      functionBuildExecutorArgs: "--insecure,--skip-tls-verify,--skip-unused-stages,--log-format=text,--cache=true"
```

## Configure function build max simultaneous jobs

You can set a custom maximum number of simultaneous jobs which can be running at the same time. The default value is set to 5.

```yaml
   spec:
      functionBuildMaxSimultaneousJobs: 5
```

## Configure healthz liveness timeout

By default, the function will be considered unhealthy if the liveness health check endpoint does not respond within 10 seconds.

```yaml
   spec:
      healthzLivenessTimeout: "10s"
```

## Configure function request body limit

This field configures the maximum size limit for the request body of a function. The default value is set to 1 megabyte.

```yaml
   spec:
      functionRequestBodyLimitMb: 1
```

## Configure function timeout

By default, the maximum execution time limit for a function is set to 180 seconds.

```yaml
   spec:
      functionTimeoutSec: 180
```

## Configure default build job preset

You can configure the default build job preset to be used. 

```yaml
   spec:
      defaultBuildJobPreset: "normal"
```

## Configure default runtime pod preset

You can configure the default runtime pod preset to be used.

```yaml
   spec:
      defaultRuntimePodPreset: "M"
```
