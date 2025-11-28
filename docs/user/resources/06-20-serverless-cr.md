# Serverless

The `serverlesses.operator.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the Serverless configuration that you want to install on your cluster. To get the up-to-date CRD and show the output in the YAML format, run this command:

   ```bash
   kubectl get crd serverlesses.operator.kyma-project.io -o yaml
   ```

## Sample Custom Resource

The following Serverless custom resource (CR) shows configuration of Serverless with custom endpoints for eventing and tracing and custom additional configuration.

   ```yaml
   apiVersion: operator.kyma-project.io/v1alpha1
   kind: Serverless
   metadata:
     finalizers:
     - serverless-operator.kyma-project.io/deletion-hook
     name: default
     namespace: kyma-system
   spec:
     eventing:
        endpoint: http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish
     tracing:
        endpoint: http://telemetry-otlp-traces.kyma-system.svc.cluster.local:4318/v1/traces
     functionRequeueDuration: 5m
     healthzLivenessTimeout: "10s"
     defaultRuntimePodPreset: "M"
     logLevel: "info"
     logFormat: "json"
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

## Custom Resource Parameters

For details, see the [Serverless specification file](https://github.com/kyma-project/serverless-manager/blob/main/components/operator/api/v1alpha1/serverless_types.go).
<!-- TABLE-START -->
### Serverless.operator.kyma-project.io/v1alpha1

**Spec:**

| Parameter                                 | Type    | Description                                                                                                                                          |
|-------------------------------------------|---------|------------------------------------------------------------------------------------------------------------------------------------------------------|
| **eventing**                              | object  |                                                                                                                                                      |
| **eventing.&#x200b;endpoint** (required)  | string  | Used Eventing endpoint                                                                                                                               |
| **tracing**                               | object  |                                                                                                                                                      |
| **tracing.&#x200b;endpoint** (required)   | string  | Used Tracing endpoint                                                                                                                                |
| **functionRequeueDuration**               | string  | Sets the requeue duration for Function. By default, the Function associated with the default configuration is requeued every 5 minutes               |
| **healthzLivenessTimeout**                | string  | Sets the timeout for the Function health check. The default value in seconds is `10`                                                                 |
| **defaultRuntimePodPreset**               | string  | Configures the default runtime Pod preset to be used                                                                                                 |
| **logLevel**                              | string  | Sets desired log level to be used. The default value is "info"                                                                                       |
| **logFormat**                             | string  | Sets desired log format to be used. The default value is "json"                                                                                      |

**Status:**

| Parameter                                            | Type       | Description                                                                                                                                                                                                                                                                                                                                                    |
|------------------------------------------------------|------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **conditions**                                       | \[\]object | Conditions associated with CustomStatus.                                                                                                                                                                                                                                                                                                                       |
| **conditions.&#x200b;lastTransitionTime** (required) | string     | Specifies the last time the condition transitioned from one status to another. This should be when the underlying condition changes.  If that is not known, then using the time when the API field changed is acceptable.                                                                                                                                      |
| **conditions.&#x200b;message** (required)            | string     | Provides a human-readable message indicating details about the transition. This may be an empty string.                                                                                                                                                                                                                                                        |
| **conditions.&#x200b;observedGeneration**            | integer    | Represents the **.metadata.generation** that the condition was set based upon. For instance, if **.metadata.generation** is currently `12`, but the **.status.conditions[x].observedGeneration** is `9`, the condition is out of date with respect to the current state of the instance.                                                                       |
| **conditions.&#x200b;reason** (required)             | string     | Contains a programmatic identifier indicating the reason for the condition's last transition. Producers of specific condition types may define expected values and meanings for this field and whether the values are considered a guaranteed API. The value should be a camelCase string. This field may not be empty.                                        |
| **conditions.&#x200b;status** (required)             | string     | Specifies the status of the condition. The value is either `True`, `False`, or `Unknown`.                                                                                                                                                                                                                                                                      |
| **conditions.&#x200b;type** (required)               | string     | Specifies the condition type in camelCase or in `foo.example.com/CamelCase`. Many **.conditions.type** values are consistent across resources like `Available`, but because arbitrary conditions can be useful (see **.node.status.conditions**), the ability to deconflict is important. The regex it matches is `(dns1123SubdomainFmt/)?(qualifiedNameFmt)`. | |
| **eventingEndpoint**                                 | string     | Used Eventing endpoint.                                                                                                                                                                                                                                                                                                                                        |
| **served** (required)                                | string     | Served signifies that current Serverless is managed. Value can be one of `True`, or `False`.                                                                                                                                                                                                                                                                   |
| **state**                                            | string     | Signifies the current state of Serverless. Value can be one of `Ready`, `Processing`, `Error`, or `Deleting`.                                                                                                                                                                                                                                                  |
| **tracingEndpoint**                                  | string     | Used Tracing endpoint.                                                                                                                                                                                                                                                                                                                                         | |
| **functionRequeueDuration**                          | string     | Used the Function requeue duration.                                                                                                                                                                                                                                                                                                                            | |
| **healthzLivenessTimeout**                           | string     | Used the healthz liveness timeout.                                                                                                                                                                                                                                                                                                                             | |
| **defaultRuntimePodPreset**                          | string     | Used the default runtime Pod preset.                                                                                                                                                                                                                                                                                                                           |
| **logLevel**                                         | string     | Used the log level.                                                                                                                                                                                                                                                                                                                                            |
| **logFormat**                                        | string     | Used the log format.                                                                                                                                                                                                                                                                                                                                           |

<!-- TABLE-END -->

### Status Reasons

Processing of a Serverless CR can succeed, continue, or fail for one of these reasons:

## Serverless CR Conditions

This section describes the possible states of the Serverless CR. Three condition types, `Installed`, `Configured`, `DeploymentFailure` and `Deleted`, are used.

| No | CR State   | Condition type    | Condition status | Condition reason         | Remark                                              |
|----|------------|-------------------|------------------|--------------------------|-----------------------------------------------------|
| 1  | Processing | Configured        | true             | Configured               | Serverless configuration verified                   |
| 2  | Processing | Configured        | unknown          | ConfigurationCheck       | Serverless configuration verification ongoing       |
| 3  | Error      | Configured        | false            | ConfigurationCheckErr    | Serverless configuration verification error         |
| 4  | Error      | Configured        | false            | ServerlessDuplicated     | Only one Serverless CR is allowed                   |
| 5  | Ready      | Installed         | true             | Installed                | Serverless workloads deployed                       |
| 6  | Processing | Installed         | unknown          | Installation             | Deploying serverless workloads                      |
| 7  | Error      | Installed         | false            | InstallationErr          | Serverless resources installation error             |
| 8  | Error      | DeploymentFailure | true             | DeploymentReplicaFailure | Serverless manager has the ReplicaFailure condition |
| 9  | Deleting   | Deleted           | unknown          | Deletion                 | Deletion in progress                                |
| 10 | Deleting   | Deleted           | true             | Deleted                  | Serverless module deleted                           |
| 11 | Error      | Deleted           | false            | DeletionErr              | Deletion failed                                     |
