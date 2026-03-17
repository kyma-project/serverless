# Cluster Roles
Cluster Roles are a set of permissions that can be applied across the entire cluster. They are used to define what actions can be performed on which resources within the cluster. Cluster Roles can be assigned to users, groups, or service accounts through Role Bindings or Cluster Role Bindings.

## Provided roles
List of all Cluster Roles provided for the user.

| Name                 | Resource   | Verbs                                           | Installed with      | Aggregated to role |
|----------------------|------------|-------------------------------------------------|---------------------|--------------------|
| kyma-serverless-edit | serverless | create, delete, get, list, patch, update, watch | serverless operator | edit               |
| kyma-functions-edit  | function   | create, delete, get, list, patch, update, watch | serverless          | edit               |
| kyma-serverless-view | serverless | get, list, watch                                | serverless operator | view               |
| kyma-functions-view  | function   | get, list, watch                                | serverless          | view               |