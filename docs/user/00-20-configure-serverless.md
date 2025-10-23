# Configuring Serverless

By default, the Serverless module comes with the default configuration. You can change the configuration using the Serverless CustomResourceDefinition (CRD), which manages Serverless custom resource (CR).

## Prerequisites

- You have the [Serverless module added](https://kyma-project.io/#/02-get-started/01-quick-install).

- You have access to Kyma dashboard. Alternatively, to use CLI instructions, you must install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl).

## Context

The Serverless module has its own operator (Serverless Operator). It watches the Serverless custom resource (CR) and reconfigures (reconciles) the Serverless workloads.

The Serverless CR is an API to configure the Serverless module. You can use it to perform the following actions:

- Override endpoint for traces collected by the Serverless Functions.
- Override endpoint for Eventing.
- Override the Function requeue duration.
- Override the healthz liveness timeout.
- Override the Function request body limit.
- Override the Function timeout.
- Override the default runtime Pod preset.
- Override the default log level.
- Override the default log format.
- Enable network policies.
- Disable buildless mode of Serverless.

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

### Configuring Trace Endpoint

By default, the Serverless operator checks if there is a trace endpoint available. If available, the detected trace endpoint is used as the trace collector URL in Functions.
If no trace endpoint is detected, Functions are configured with no trace collector endpoint.
You can configure a custom trace endpoint so that Function traces are sent to any tracing backend you choose.
The currently used trace endpoint is visible in the Serverless CR status.

   ```yaml
   spec:
     tracing:
       endpoint: http://jaeger-collector.observability.svc.cluster.local:4318/v1/traces
   ```

### Configuring Eventing Endpoint

You can configure a custom Eventing endpoint to publish events sent from your Functions.
The currently used trace endpoint is visible in the Serverless CR status.
By default `http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish` is used.

   ```yaml
   spec:
     eventing:
       endpoint: http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish
   ```

### Configuring the Function Requeue Duration

By default, the Function associated with the default configuration will be requeued every 5 minutes.  

```yaml
   spec:
      functionRequeueDuration: 5m
```

### Configuring the healthz Liveness Timeout

By default, Function is considered unhealthy if the liveness health check endpoint does not respond within 10 seconds.

```yaml
   spec:
      healthzLivenessTimeout: "10s"
```

### Configuring the Default Runtime Pod Preset

You can configure the default runtime Pod preset to be used.

```yaml
   spec:
      defaultRuntimePodPreset: "M"
```

For more information on presets, see [Available Presets](https://kyma-project.io/#/serverless-manager/user/technical-reference/07-80-available-presets).

### Configuring the Log Level

You can configure the desired log level to be used.

```yaml
   spec:
      logLevel: "debug"
```

### Configuring the Log Format

You can configure the desired log format to be used.

```yaml
   spec:
      logFormat: "yaml"
```

### Enabling Network Policies

You can enable built-in network policies to ensure that the necessary communication channels required by Serverless workloads remain functional,
even on Kubernetes clusters where strict "deny-all" network policies are enforced. This allows Serverless components to operate correctly
by permitting essential traffic while maintaining a secure cluster environment.

```yaml
   spec:
      enableNetworkPolicies: true
```

### Disabling Buildless Mode

> [!NOTE]
> Buildless mode is enabled by default. To use the legacy image-building Serverless functionality, disable buildless mode through an annotation.

> [!WARNING]  
> The legacy image-building mode is deprecated and will be removed in a future version of Serverless. This functionality is scheduled for removal and will no longer be available in upcoming releases.

You can disable buildless mode of Serverless to enable the legacy image build step for Functions.

   ```yaml
    annotations:
       serverless.kyma-project.io/buildless-mode: "disabled"
   ```
