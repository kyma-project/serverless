---
name: Deploy Function from network-restricted Git repository via GitHub Actions
description: Design spec for a GitHub issue and documentation how-to guide covering CI/CD-based Function deployment when the Kyma cluster cannot reach the Git repository
type: project
---

# Deploy a Function from a Network-Restricted Git Repository via GitHub Actions

## Problem

Git-sourced Functions require the Kyma cluster to pull source code from the Git repository on every reconciliation. This fails silently when the repository is hosted on a GitHub Enterprise instance that is not reachable from the cluster network (for example, behind a corporate firewall or VPN).

The existing documentation warns about this limitation in `docs/user/tutorials/01-11-create-git-function.md` but provides no alternative path for developers in this situation.

## Proposed Solution

Document a CI/CD-based workaround: instead of letting the Kyma cluster pull from Git, a GitHub Actions workflow pushes the Function source code to the cluster as an **inline Function** on every relevant push. The cluster does not need to reach the Git repository at all.

This works because:
- The GitHub Actions runner has outbound internet access (can install Kyma CLI from github.com)
- The runner authenticates to the Kyma cluster using a static kubeconfig stored as a GitHub Secret
- `kyma function create` applies the Function inline — no Git connectivity required from the cluster side

## Scope

### In scope
- New tutorial: `docs/user/tutorials/01-12-deploy-function-from-network-restricted-git.md`
- Link from the existing warning in `01-11-create-git-function.md`

### Out of scope
- Air-gapped runners (runner also has no internet access)
- GitLab CI or other CI systems
- OIDC federation (requires cluster-to-GitHub network access for JWKS verification — not viable in this scenario)

## GitHub Issue Draft

**Title:** `docs: How to deploy Functions from a network-restricted Git repository using GitHub Actions`

**Body:**

### Problem

Git-sourced Functions require the Kyma cluster to reach the Git repository over the network. When the repository is hosted on a Git server that is not reachable from the Kyma cluster network (for example, a GitHub Enterprise instance behind a corporate firewall), the Function controller cannot fetch the source code. The Function never reaches the `RUNNING` state.

The existing documentation warns about this in the [Create a Git Function](../tutorials/01-11-create-git-function.md) tutorial but does not describe an alternative.

### Proposed solution

Add a new tutorial that documents how to use a GitHub Actions workflow to deploy a Function as an inline Function directly to the Kyma cluster. The workflow triggers on push, installs Kyma CLI using the [`kyma-project/setup-kyma-cli`](https://github.com/kyma-project/setup-kyma-cli) action, and applies the Function using a kubeconfig stored as a GitHub Secret.

### Acceptance criteria

- [ ] New file `docs/user/tutorials/01-12-deploy-function-from-network-restricted-git.md` created
- [ ] The tutorial covers: prerequisites, creating a least-privilege kubeconfig, storing it as a GitHub Secret, and a working example workflow
- [ ] The existing warning in `01-11-create-git-function.md` links to the new tutorial
- [ ] The new tutorial follows the [Kyma content guidelines](https://github.com/kyma-project/community/blob/main/docs/guidelines/content-guidelines/agentic.md)

---

## Documentation Draft

**File:** `docs/user/tutorials/01-12-deploy-function-from-network-restricted-git.md`

---

### Deploy a Function from a Network-Restricted Git Repository

> [!NOTE]
> Use this approach when your Git repository is hosted on a server that is not reachable from the Kyma cluster network, for example, a GitHub Enterprise instance behind a corporate firewall. If your cluster can reach the repository, use [Git-sourced Functions](./01-11-create-git-function.md) instead.

This tutorial shows how to use a GitHub Actions workflow to deploy a Kyma Function whenever you push changes to your repository. Instead of letting the cluster pull source code from Git, the workflow pushes the Function code directly to the cluster as an inline Function. The cluster does not need to reach the Git repository.

#### Prerequisites

- You have the [Serverless module added](https://kyma-project.io/02-get-started/01-quick-install.html).
- You have a GitHub repository with GitHub Actions enabled.
- You have `kubectl` and [Kyma CLI](https://kyma-project.io/external-content/cli/docs/user/README.html) installed locally.
- You have cluster admin access to generate a kubeconfig.

#### How It Works

On every push to the configured branch, GitHub Actions:

1. Checks out your repository code.
2. Installs Kyma CLI using [`kyma-project/setup-kyma-cli`](https://github.com/kyma-project/setup-kyma-cli).
3. Authenticates to the Kyma cluster using a kubeconfig stored as a GitHub Secret.
4. Deploys the Function as an inline Function using `kyma function create`.

#### Procedure

**Step 1: Create a service account and kubeconfig**

Create a dedicated Kubernetes service account with the minimum permissions required to manage Functions in your namespace.

```bash
export NAMESPACE={YOUR_NAMESPACE}

kubectl -n $NAMESPACE create serviceaccount function-deployer

kubectl -n $NAMESPACE create role function-deployer-role \
  --verb=get,list,create,update,patch,delete \
  --resource=functions

kubectl -n $NAMESPACE create rolebinding function-deployer-binding \
  --role=function-deployer-role \
  --serviceaccount=$NAMESPACE:function-deployer
```

Generate a kubeconfig for the service account:

```bash
kubectl -n $NAMESPACE create token function-deployer --duration=8760h \
  | xargs -I TOKEN kubectl config set-credentials function-deployer --token=TOKEN

# Export the kubeconfig to a file
kubectl config view --minify --flatten > function-deployer-kubeconfig.yaml
```

> [!WARNING]
> The generated token grants write access to Functions in the namespace. Store it securely and rotate it regularly.

**Step 2: Store the kubeconfig as a GitHub Secret**

1. Open your GitHub repository and go to **Settings** > **Secrets and variables** > **Actions**.
2. Select **New repository secret**.
3. Set the name to `KYMA_KUBECONFIG` and paste the contents of `function-deployer-kubeconfig.yaml` as the value.
4. Select **Add secret**.

**Step 3: Add the GitHub Actions workflow**

Create the file `.github/workflows/deploy-function.yml` in your repository:

```yaml
name: Deploy Kyma Function

on:
  push:
    branches: [main]
    paths:
      - 'function/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: kyma-project/setup-kyma-cli@v1
        with:
          version: stable

      - name: Deploy Function
        env:
          KUBECONFIG_CONTENT: ${{ secrets.KYMA_KUBECONFIG }}
        run: |
          echo "$KUBECONFIG_CONTENT" > /tmp/kubeconfig.yaml
          KUBECONFIG=/tmp/kubeconfig.yaml kyma function create my-function \
            --namespace {YOUR_NAMESPACE} \
            --runtime nodejs24 \
            --source function/handler.js \
            --dependencies function/package.json
```

> [!NOTE]
> Adjust `paths`, `branches`, `--namespace`, `--runtime`, `--source`, and `--dependencies` to match your repository structure and runtime.

**Step 4: Verify the deployment**

After pushing to the `main` branch, check that the Function reached the `RUNNING` state:

```bash
kubectl get functions my-function -n {YOUR_NAMESPACE}
```

Expected output:

```bash
NAME          CONFIGURED   BUILT   RUNNING   RUNTIME    VERSION   AGE
my-function   True         True    True      nodejs24   1         60s
```

#### Security Considerations

- Use a **dedicated service account** with the minimum required permissions. Do not use a cluster admin kubeconfig.
- **Rotate the token** regularly. The `kubectl create token` command accepts a `--duration` flag; set a rotation schedule that matches your security policy.
- The kubeconfig grants write access to Functions in the namespace — treat it as a sensitive credential.
