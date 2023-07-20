---
title: Serverless
---

The `serverlesses.operator.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the Serverless configuration that you want to install on your cluster. To get the up-to-date CRD and show the output in the YAML format, run this command:

   ```bash
   kubectl get crd serverlesses.operator.kyma-project.io -o yaml
   ```

## Sample custom resource

The following Serverless custom resource (CR) shows configuration of Serverless with the external registry and two user URLs for Publisher Proxy and Trace Collector.

   ```yaml
   apiVersion: operator.kyma-project.io/v1alpha1
   kind: Serverless
   metadata:
     finalizers:
     - serverless-operator.kyma-project.io/deletion-hook
     name: default
     namespace: kyma-system
   spec:
     dockerRegistry:
       enableInternal: false
       secretName: my-secret
     eventPublisherProxyURL: http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish
     traceCollectorURL: http://telemetry-otlp-traces.kyma-system.svc.cluster.local:4318/v1/traces
   status:
     conditions:
     - lastTransitionTime: "2023-04-28T10:09:37Z"
       message: Configured with default Publisher Proxy URL and default Trace Collector
         URL.
       reason: Configured
       status: "True"
       type: Configured
     - lastTransitionTime: "2023-04-28T10:15:15Z"
       message: Serverless installed
       reason: Installed
       status: "True"
       type: Installed
     eventPublisherProxyURL: http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish
     state: Ready
     traceCollectorURL: http://telemetry-otlp-traces.kyma-system.svc.cluster.local:4318/v1/traces
   ```

## Custom resource parameters

For details, see the [Serverless configuration file](https://github.com/kyma-project/serverless-manager/blob/main/api/v1alpha1/serverless_types.go).

| Parameter         | Description                                   |
| ---------------------------------------- | ---------|
| **spec.eventPublisherProxyURL** | Event publisher endpoint used by [the Serverless SDK](https://kyma-project.io/docs/kyma/latest/05-technical-reference/svls-08-function-specification/#event-object-sdk). |
| **spec.traceCollectorURL** | URL for your Trace Collector.  |
| **spec.dockerRegistry** | Specifies the configuration of the registry used to store the images of the built functions. |
| **spec.dockerRegistry.enableInternal** | If set to `true`, the internal Serverless Docker registry is used. If set to `false`, provide `secretName` which must be in the same Namespace as Serverless CR. This Secret must contain your address and password to the external Docker registry. If you don't provide your `secretName`, Serverless assumes that installation is provided on k3d and it expects the up-and-running k3d registry. |
| **spec.dockerRegistry.secretName** | Includes the address and credentials to the external Docker registry. |
| **status.eventPublisherProxyURL** | The eventing endpoint used for the installation. |
| **status.traceCollector** | The tracing endpoint used for the installation. |
| **status.state** | The possible transition types are:<br>- `Ready`: The instance is ready and usable.<br>- `Processing`: The instance is being installed or configured. <br>- `Error`: The operation cannot be executed. <br>- `Delete`: The instance is being deleted. |
| **status.conditions** | An array of conditions describing the status of Serverless. |
| **status.conditions.reason** | An array of conditions describing the status of Serverless. |
| **status.conditions.type** | The possible transition types are:<br>- `Installed`: The instance is ready and usable.<br>- `Configured`: The instance is configured. |
<!-- TABLE-END -->