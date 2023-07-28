# Overview

Following the motivation of [Kyma modularization](https://kyma-project.io/docs/kyma/latest/01-overview/#kyma-modules), Serverless is now a module that you can enable and disable. 
Serverless has its own Serverless Operator which is installed in the target Kyma runtime by infrastructure operators based on the module descriptor (`Module Template`).
Serverless Operator watches Serverless custom resource (CR) to re-configure (reconcile) the Serverless instalation.

## Serverless CR

Reconciliation of the Serverless components is driven by the content of the Serverless CR.

You can use Serverless CR for the Serverless configuration with the provided API, for example:
 - override trace endpoint
 - override eventing endpoint
 - enable/disable internal docker registry
 - configure external docker registry

You can also see the status of the Serverless module using Serverless CR, for example:
 - health of the Serverless workloads (for example, controller, webhook, Docker registry)
 - URL of the detected event publisher proxy
 - URL of the detected/configured OpenTelemetry protocol (OTLP) endpoints
 - indication whether internal Docker registry is used or URL of the configured Docker registry

   ```yaml
   apiVersion: operator.kyma-project.io/v1alpha1
   kind: Serverless
     name: serverless-sample
   spec:
     dockerRegistry:
       enableInternal: true
       secretName: xxxx 
     eventingPublisherProxy: http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish
     otlpTracesEndpoint: http://tracing-jaeger-collector.kyma-system.svc.cluster.local:2342/v1/metrics ##<-- this is a dummy example
     otlpMetricsEndpoint: http://tracing-jaeger-collector.kyma-system.svc.cluster.local:4318/v1/trace
     defaultFunctionRuntimePreset: M
     defaultFunctionBuildPreset: S
     maxParallelFunctionBuilds: 5
     controllerLogLevel: debug
     webhookLogLevel: debug
     ## runtime config
     maxRequestPayloadSize: 2MB
     functionTimeoutSeconds: 180
   status:
    # health of serverless workloads (i.e controller, webhook, docker registry installed)
    # url of the detected event publisher proxy
    # url of the detected/configured otlp endpoints
    # indication whether internal docker registry is used / url of configured docker registry
   ```

## Dependencies

There are other Kyma modules that are watched by Serverless Operator:
 - Eventing (soft dependency)
 - Telemetry (soft dependency)

The Eventing and Telementry modules are not required to install Serverless module. If they are discovered they deliver default values for the Serverless configuration (for example, `publisherProxyEndpoint`, `otlpEnpoints`).
The detected and used endpoints must be part of the Serverless CR status.


![deps](../assets/modular-serverless.drawio.svg)

>NOTE: Until the dependant modules are discoverable in the modular fashion the Serverless Operator can test the availibility of the endpoints directly.

## Propagating configuration to the Function runtime

When a dependant module is discovered or when you override the Serverless CR manually, the Serverless Operator reconciles the Serverless Controller.
If the change affect the Function runtime (for example, `otlpEndpoint`), the Serverless Controller automatically changes the ENVs in the Pod templates of the Function deployments. This restarts the Functions and new ENV values are consumed.