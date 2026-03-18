# Cluster Roles
ClusterRoles are a set of permissions that can be applied across the entire cluster. They define which actions can be performed on which resources within the cluster. ClusterRoles can be assigned to users, groups, or service accounts through RoleBindings or ClusterRoleBindings.

## Provided Roles
List of all ClusterRoles provided for the user.

| Name                 | Resource   | Verbs                                           | Installed with      | Aggregated to role |
|----------------------|------------|-------------------------------------------------|---------------------|--------------------|
| kyma-serverless-edit | serverless | create, delete, get, list, patch, update, watch | serverless operator | edit               |
| kyma-functions-edit  | function   | create, delete, get, list, patch, update, watch | serverless          | edit               |
| kyma-serverless-view | serverless | get, list, watch                                | serverless operator | view               |
| kyma-functions-view  | function   | get, list, watch                                | serverless          | view               |