# Network Policies

## Overview

The Serverless module defines network policies to ensure communication within the Kubernetes cluster, particularly in environments where a deny-all network policy is applied.

When a cluster-wide deny-all network policy is enforced, which blocks all ingress and egress traffic by default, the Serverless network policies explicitly allow only the necessary communication paths to ensure the module functions correctly.

## Network Policies

To list the network policies belonging to the Serverless module, run the following command:

```bash
kubectl get networkpolicies -n kyma-system -l kyma-project.io/module=serverless
```

The following tables describe the network policies for the Serverless module.

**Serverless Policies**

| Policy Name | Type | Port(s) | Description |
|---|---|---|---|
| `kyma-project.io--serverless-allow-egress` | Egress | All | Allows unrestricted outbound traffic from Pods labeled `networking.serverless.kyma-project.io/from-serverless: allowed`. This is used by the Serverless controller to fetch Function source code from external Git repositories. |
| `kyma-project.io--serverless-allow-metrics` | Ingress | 8080 (TCP) | Allows ingress to the metrics endpoint from Pods labeled `app.kubernetes.io/instance: rma` or `networking.kyma-project.io/metrics-scraping: allowed` for metrics scraping. Applied to all Pods labeled `kyma-project.io/module: serverless`. |
| `kyma-project.io--serverless-allow-to-apiserver` | Egress | 443 (TCP) | Allows egress from Pods labeled `networking.serverless.kyma-project.io/to-apiserver: allowed` to the Kubernetes API server. |
| `kyma-project.io--serverless-allow-to-dns` | Egress | 53 (TCP/UDP), 8053 (TCP/UDP) | Allows egress to DNS services for cluster and external DNS resolution. Targets any IP on port 53, and Pods labeled `k8s-app: kube-dns` or `k8s-app: node-local-dns` in the `kube-system` namespace on ports 53 and 8053. Applied to all Pods labeled `kyma-project.io/module: serverless`. |

**Serverless Operator Policies**

| Policy Name | Type | Port(s) | Description |
|---|---|---|---|
| `kyma-project.io--serverless-operator-allow-to-apiserver` | Egress | 443 (TCP), 6443 (TCP) | Allows egress from the Serverless Operator Pod to the Kubernetes API server. |
| `kyma-project.io--serverless-operator-allow-to-dns` | Egress | 53 (TCP/UDP), 8053 (TCP/UDP) | Allows egress from the Serverless Operator Pod to DNS services for cluster and external DNS resolution. Targets any IP on port 53, and Pods labeled `k8s-app: kube-dns` or `k8s-app: node-local-dns` in the `kube-system` namespace on ports 53 and 8053. |
