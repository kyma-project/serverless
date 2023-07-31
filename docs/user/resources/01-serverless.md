---
title: Serverless
---

The `serverlesses.operator.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the Serverless configuration that you want to install on your cluster. To get the up-to-date CRD and show the output in the YAML format, run this command:

   ```bash
   kubectl get crd serverlesses.operator.kyma-project.io -o yaml
   ```

## Sample custom resource

The following Serverless custom resource (CR) shows configuration of Serverless with the external registry and custom endpoints for eventing and tracing.

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
     eventing:
        endpoint: http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish
     tracing:
        endpoint: http://telemetry-otlp-traces.kyma-system.svc.cluster.local:4318/v1/traces
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
<!-- TABLE-START -->
### Serverless.operator.kyma-project.io/v1alpha1

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **dockerRegistry**  | object |  |
| **dockerRegistry.&#x200b;enableInternal**  | boolean |  |
| **dockerRegistry.&#x200b;secretName**  | string |  |
| **eventing**  | object |  |
| **eventing.&#x200b;endpoint** (required) | string |  |
| **tracing**  | object |  |
| **tracing.&#x200b;endpoint** (required) | string |  |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **conditions**  | \[\]object | Conditions associated with CustomStatus. |
| **conditions.&#x200b;lastTransitionTime** (required) | string | lastTransitionTime is the last time the condition transitioned from one status to another. This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable. |
| **conditions.&#x200b;message** (required) | string | message is a human readable message indicating details about the transition. This may be an empty string. |
| **conditions.&#x200b;observedGeneration**  | integer | observedGeneration represents the .metadata.generation that the condition was set based upon. For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date with respect to the current state of the instance. |
| **conditions.&#x200b;reason** (required) | string | reason contains a programmatic identifier indicating the reason for the condition's last transition. Producers of specific condition types may define expected values and meanings for this field, and whether the values are considered a guaranteed API. The value should be a CamelCase string. This field may not be empty. |
| **conditions.&#x200b;status** (required) | string | status of the condition, one of True, False, Unknown. |
| **conditions.&#x200b;type** (required) | string | type of condition in CamelCase or in foo.example.com/CamelCase. --- Many .condition.type values are consistent across resources like Available, but because arbitrary conditions can be useful (see .node.status.conditions), the ability to deconflict is important. The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt) |
| **dockerRegistry**  | string | Used registry configuration. Contains registry URL or "internal" |
| **eventingEndpoint**  | string | Used the Eventing endpoint and the Tracing endpoint. |
| **served** (required) | string | Served signifies that current Serverless is managed. Value can be one of ("True", "False"). |
| **state**  | string | State signifies current state of Serverless. Value can be one of ("Ready", "Processing", "Error", "Deleting"). |
| **tracingEndpoint**  | string |  |

<!-- TABLE-END -->

### Status reasons

Processing of a Serverless CR can succeed, continue, or fail for one of these reasons:


//TODO Provide description of Serverless CR conditions here