# New Function Replica is Not Starting

## Symptom

After migration to serverless `1.9.3` some of the Functions show `Running=False` condition.

```bash
kubectl get function -A

NAMESPACE   NAME                        CONFIGURED   RUNNING   RUNTIME    VERSION   AGE
foo         function1                   True         True      nodejs22   2         2d
foo         function2                   True         True      nodejs22   3         2d
foo         function3                   True         False     nodejs22   2         2d
foo         function4                   True         False     nodejs22   3         2d
```

## Cause

In Serverless 1.9.3 we introduced [buildless mode](../00-60-buildless-serverless.md), which removes the in-cluster image build step. This reduces overall resource usage and speeds up delivery. As a result, dependency resolution (`pip install`/`npm install`) now happens at Function pod startup. During this brief initialization phase, the pod may require slightly more CPU and memory. If the Function’s resource limits are very low (for example, custom defined strict memory/CPU limits via `resourceConfiguration`), the pod can be OOMKilled by Kubernetes. 

## Solution

To avoid this, consider increasing the resource limits in the `spec.resourceConfiguration.function` or using a [larger preset](../technical-reference/07-80-available-presets.md), especially for Functions with multiple or heavy dependencies.

```yaml
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  labels:
    app.kubernetes.io/name: function3
  name: function3
  namespace: foo
  uid: ...
spec:
...
  resourceConfiguration:
    function:
      resources:
        limits:
          cpu: 100m # needs increasing
          memory: 64Mi
        requests:
          cpu: 50m # needs increasing
          memory: 32Mi
```

Learn more how to specify resources per Function [here](../resources/06-10-function-cr.md)
Learn more how to configure default resource preset for all functions [here](../00-20-configure-serverless.md)
