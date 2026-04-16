# Cluster Roles

The Serverless module ships a set of ClusterRoles that extend the standard Kubernetes `view` and `edit` aggregated roles. When these ClusterRoles are installed, any user or service account already bound to the built-in `view` or `edit` role automatically gains the corresponding permissions for Serverless resources — no additional RoleBindings are required.

## Provided ClusterRoles

### Function Roles

They are installed with the Serverless module. They cover the `Function` custom resource (CR) from the `serverless.kyma-project.io` API group.

| Name | Aggregated to | Resources | Verbs |
|---|---|---|---|
| `kyma-functions-edit` | `edit` | `functions` | `create`, `delete`, `get`, `list`, `patch`, `update`, `watch` |
| `kyma-functions-edit` | `edit` | `functions/status` | `get` |
| `kyma-functions-edit` | `edit` | `functions/scale` | `get`, `patch`, `update` |
| `kyma-functions-view` | `view` | `functions` | `get`, `list`, `watch` |
| `kyma-functions-view` | `view` | `functions/status` | `get` |
| `kyma-functions-view` | `view` | `functions/scale` | `get` |

### Serverless Operator Roles

They are installed with the Serverless Operator. They cover the `Serverless` CR from the `operator.kyma-project.io` API group, which is used to configure the module itself.

| Name | Aggregated to | Resources | Verbs |
|---|---|---|---|
| `kyma-serverless-edit` | `edit` | `serverlesses` | `create`, `delete`, `get`, `list`, `patch`, `update`, `watch` |
| `kyma-serverless-edit` | `edit` | `serverlesses/status` | `get` |
| `kyma-serverless-view` | `view` | `serverlesses` | `get`, `list`, `watch` |
| `kyma-serverless-view` | `view` | `serverlesses/status` | `get` |

## Verify Installed Roles

To list all ClusterRoles provided by the Serverless module, run:

```bash
kubectl get clusterroles -l kyma-project.io/module=serverless
```
