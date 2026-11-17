# Deploy a Function from a Network-Restricted Git Repository

> [!NOTE]
> Use this approach when your Git repository is hosted on a server that is not reachable from the Kyma cluster network, for example, a GitHub Enterprise instance behind a corporate firewall. If your cluster can reach the repository, use [Git-sourced Functions](./01-11-create-git-function.md) instead.

This tutorial shows how to use a GitHub Actions workflow to deploy a Kyma Function whenever you push changes to your repository. Instead of letting the cluster pull source code from Git, the workflow pushes the Function code directly to the cluster as an inline Function. The cluster does not need to reach the Git repository.

## Prerequisites

- You have the [Serverless module added](https://kyma-project.io/02-get-started/01-quick-install.html).
- You have a GitHub repository with GitHub Actions enabled.
- You have `kubectl` installed locally.
- You have cluster admin access to create service accounts and roles.

## How It Works

On every push to the configured branch, GitHub Actions:

1. Checks out your repository code.
2. Installs Kyma CLI using [`kyma-project/setup-kyma-cli`](https://github.com/kyma-project/setup-kyma-cli).
3. Authenticates to the Kyma cluster using a kubeconfig stored as a GitHub Secret.
4. Deploys the Function as an inline Function using `kyma function create nodejs`.

> [!NOTE]
> The `kyma function` commands are Kyma CLI extensions served by the Serverless module installed on your cluster. The GitHub Actions runner must be able to reach the Kyma API server to discover and run these commands.

## Procedure

### Step 1: Create a Service Account and Kubeconfig

Create a dedicated Kubernetes service account with the minimum permissions required to manage Functions in your namespace.

```bash
export NAMESPACE={YOUR_NAMESPACE}

kubectl -n $NAMESPACE create serviceaccount function-deployer

kubectl -n $NAMESPACE create role function-deployer-role \
  --verb=get,list,create,update,patch,delete \
  --resource=functions.serverless.kyma-project.io

kubectl -n $NAMESPACE create rolebinding function-deployer-binding \
  --role=function-deployer-role \
  --serviceaccount=$NAMESPACE:function-deployer
```

Generate a token and save it as a kubeconfig file:

```bash
TOKEN=$(kubectl -n $NAMESPACE create token function-deployer --duration=8760h)

CLUSTER=$(kubectl config view --minify -o jsonpath='{.clusters[0].name}')
kubectl config set-credentials function-deployer --token=$TOKEN
kubectl config set-context function-deployer-ctx \
  --cluster=$CLUSTER \
  --user=function-deployer \
  --namespace=$NAMESPACE

CURRENT_CONTEXT=$(kubectl config current-context)
kubectl config use-context function-deployer-ctx
kubectl config view --minify --flatten > function-deployer-kubeconfig.yaml

# Restore your original context
kubectl config use-context $CURRENT_CONTEXT
```

> [!WARNING]
> The generated token may expire earlier than requested. Managed Kyma clusters cap the token lifetime (often 24–48 hours). If the workflow starts failing with authentication errors, the token has expired — re-run Step 1 to generate a new one and update the GitHub Secret. Store the kubeconfig securely and rotate it regularly.

### Step 2: Store the Kubeconfig as a GitHub Secret

1. Open your GitHub repository and go to **Settings** > **Secrets and variables** > **Actions**.
2. Select **New repository secret**.
3. Set the name to `KYMA_KUBECONFIG` and paste the contents of `function-deployer-kubeconfig.yaml` as the value.
4. Select **Add secret**.

### Step 3: Add the GitHub Actions Workflow

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
          kyma --kubeconfig /tmp/kubeconfig.yaml function create nodejs my-function \
            --namespace {YOUR_NAMESPACE} \
            --source function/handler.js \
            --dependencies function/package.json
```

> [!NOTE]
> Replace `{YOUR_NAMESPACE}` and `my-function` with your actual namespace and Function name. Adjust `paths`, `branches`, `--source`, and `--dependencies` to match your repository structure. To use Python, replace `nodejs` with `python`. For a specific Node.js version, add `--version 22` or `--version 26`.

### Step 4: Verify the Deployment

After pushing to the `main` branch, check that the Function reached the `RUNNING` state:

```bash
kubectl get functions my-function -n {YOUR_NAMESPACE}
```

Expected output:

```text
NAME          CONFIGURED   BUILT   RUNNING   RUNTIME    VERSION   AGE
my-function   True         True    True      nodejs24   1         60s
```

## Security Considerations

- Use a **dedicated service account** with the minimum required permissions. Don't use a cluster admin kubeconfig.
- **Rotate the token regularly.** The `kubectl create token` command accepts a `--duration` flag. Set a rotation schedule that matches your security policy.
- The kubeconfig grants write access to Functions in the namespace. Treat it as a sensitive credential.
