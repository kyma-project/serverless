# Overview

Following the motivation of [Kyma modularization](https://kyma-project.io/docs/kyma/latest/01-overview/#kyma-modules), Serverless is now a module that you can enable and disable. 
Serverless has its own Serverless Operator, which is installed in the target Kyma runtime by infrastructure operators based on the module descriptor (`Module Template`).
Serverless Operator watches the Serverless custom resource (CR) to reconfigure (reconcile) the Serverless installation.

## Serverless CR

Reconciliation of the Serverless components is driven by the content of the Serverless CR.

You can use the Serverless CR for the Serverless configuration with the provided API, for example:
 - override trace endpoint
 - override Eventing endpoint
 - enable/disable internal Docker registry
 - configure external Docker registry

You can also see the status of the Serverless module using Serverless CR, for example:
 - health of the Serverless workloads (for example, controller, webhook, Docker registry)
 - URL of the currently configured Eventing endpoint
 - URL of the currently configured trace endpoint
 - indication whether an internal Docker registry is used, or the URL of the configured Docker registry

   ```yaml
   apiVersion: operator.kyma-project.io/v1alpha1
   kind: Serverless
     name: serverless-sample
   spec:
     dockerRegistry:
       enableInternal: false
       secretName: xxxx 
     tracing: 
        endpoint: http://tracing-jaeger-collector.kyma-system.svc.cluster.local:2342/v1/metrics
     eventing: 
        endpoint: http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish
   status:
    # health of serverless workloads (i.e controller, webhook, docker registry installed)
    # url of the detected event publisher proxy
    # url of the detected/configured otlp endpoints
    # indication whether internal docker registry is used / url of configured docker registry
   ```

## Dependencies

Serverless Operator also watches other Kyma modules, such as:
 - Telemetry (soft dependency)

You don't need to install the Telemetry module in order to install the Serverless module. But if the Telemetry module is identified, it delivers default values for the trace endpoint in the Serverless configuration.
The detected and used endpoints are a part of the Serverless CR status.


![deps](../assets/modular-serverless.drawio.svg)

> **NOTE:** Until the dependent modules are discoverable in the modular way, the Serverless Operator can test the availability of the endpoints directly.

## Propagating configuration to the Function runtime

When a dependent module is discovered or when you override the Serverless CR manually, Serverless Operator reconciles Serverless Controller.
If the change affects the Function runtime (for example, `otlpEndpoint`), Serverless Controller automatically changes the environment values in the Pod templates of the Function deployments. This restarts the Functions, and new environment values are consumed.