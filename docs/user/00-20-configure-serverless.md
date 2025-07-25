# Serverless Configuration

## Overview

The Serverless module has its own operator (Serverless Operator). It watches the Serverless custom resource (CR) and reconfigures (reconciles) the Serverless workloads.

The Serverless CR is an API to configure the Serverless module. You can use it to perform the following actions:

- Override endpoint for traces collected by the Serverless Functions
- Override endpoint for Eventing
- Override the Function requeue duration
- Override the healthz liveness timeout
- Override the Function request body limit
- Override the Function timeout
- Override the default runtime Pod preset
- Override the default log level
- Override the default log format

The default configuration of the Serverless Module is following:

   ```yaml
   apiVersion: operator.kyma-project.io/v1alpha1
   kind: Serverless
   metadata:
     name: serverless-sample
   spec:
     dockerRegistry:
       enableInternal: true
   ```

## Configure Trace Endpoint

By default, the Serverless operator checks if there is a trace endpoint available. If available, the detected trace endpoint is used as the trace collector URL in Functions.
If no trace endpoint is detected, Functions are configured with no trace collector endpoint.
You can configure a custom trace endpoint so that Function traces are sent to any tracing backend you choose.
The currently used trace endpoint is visible in the Serverless CR status.

   ```yaml
   spec:
     tracing:
       endpoint: http://jaeger-collector.observability.svc.cluster.local:4318/v1/traces
   ```

## Configure Eventing Endpoint

You can configure a custom Eventing endpoint to publish events sent from your Functions.
The currently used trace endpoint is visible in the Serverless CR status.
By default `http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish` is used.

   ```yaml
   spec:
     eventing:
       endpoint: http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish
   ```

## Configure the Function Requeue Duration

By default, the Function associated with the default configuration will be requeued every 5 minutes.  

```yaml
   spec:
      functionRequeueDuration: 5m
```

## Configure the healthz Liveness Timeout

By default, Function is considered unhealthy if the liveness health check endpoint does not respond within 10 seconds.

```yaml
   spec:
      healthzLivenessTimeout: "10s"
```

## Configure the Default Runtime Pod Preset

You can configure the default runtime Pod preset to be used.

```yaml
   spec:
      defaultRuntimePodPreset: "M"
```

For more information on presets, see [Available Presets](https://kyma-project.io/#/serverless-manager/user/technical-reference/07-80-available-presets).

## Configure the Log Level

You can configure the desired log level to be used.

```yaml
   spec:
      logLevel: "debug"
```

## Configure the Log Format

You can configure the desired log format to be used.

```yaml
   spec:
      logFormat: "yaml"
```

## Enable Network Policies

You can enable built-in network policies to ensure that the necessary communication channels required by Serverless workloads remain functional,
even on Kubernetes clusters where strict "deny-all" network policies are enforced. This allows Serverless components to operate correctly
by permitting essential traffic while maintaining a secure cluster environment.

```yaml
   spec:
      enableNetworkPolicies: true
```