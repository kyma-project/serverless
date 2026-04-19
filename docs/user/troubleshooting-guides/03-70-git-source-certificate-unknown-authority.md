# Git Source Certificate Signed by Unknown Authority

## Symptom

When using a Git repository as a Function source over HTTPS, the Function status shows a condition with reason `SourceUpdateFailed` and a message similar to:

```bash
certificate signed by unknown authority
```

The Function remains in a non-ready state and its source code cannot be fetched or updated.

## Cause

The Serverless manager cannot verify the TLS certificate of the Git repository server because the certificate is signed by a Certificate Authority (CA) that is not trusted in the cluster.

This typically happens when:

- The Git server uses a certificate signed by a private or internal CA (for example, an on-premise GitLab or GitHub Enterprise instance).
- The cluster's root CA bundle does not include the CA that signed the Git server's certificate.
- The cluster or landscape root CAs are managed externally and are not under the control of the Serverless module.

## Solution

The root CA bundle used by the cluster must include the CA that signed your Git server's certificate. This is a cluster-level or landscape-level configuration — it cannot be adjusted within the Serverless module itself.

Depending on your environment, take one of the following steps:

- **SAP BTP, Kyma runtime**: Contact your landscape administrator to ensure the required root CA is included in the cluster's trusted CA bundle.
- **Self-managed cluster**: Add the CA certificate to the cluster's system trust store or configure the relevant node/Pod CA bundle to include the missing CA. Refer to your Kubernetes distribution documentation for the exact procedure.
- **Testing only**: If you want to skip TLS verification for a specific Git repository (not recommended for production), set the `spec.source.gitRepository.auth` field appropriately — see [Use Git Sources in Functions](../technical-reference/07-40-git-source-type.md) for details.

After the CA bundle is updated, the Serverless manager will automatically retry fetching the Git source and the Function status should return to `Ready`.
