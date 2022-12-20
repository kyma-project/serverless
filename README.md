# Serverless Manager

## Overview

Serverless Manager allows deploying the [Serverless](https://kyma-project.io/docs/kyma/latest/01-overview/main-areas/serverless/) component on the Kyma cluster in compatibility with the [Lifecycle Manager](https://github.com/kyma-project/lifecycle-manager).

## Prerequisites

- Access to a k8s cluster
- [Go](https://go.dev/)
- [k3d](https://k3d.io/)
- [Docker](https://www.docker.com/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Kubebuilder](https://book.kubebuilder.io/)

## Installation on the k3d cluster

1. Clone the project.

    ```bash
    git clone https://github.com/kyma-project/serverless-manager.git && cd serverless-manager/
    ```

2. Build the manager locally and run it on the k3d cluster.

    ```bash
    make k3d-run
    ```

> **NOTE:** To clean up the k3d cluster, use the `make k3d-stop` make target.

## Manual installation on the k3d cluster

1. Clone the project.

    ```bash
    git clone https://github.com/kyma-project/serverless-manager.git && cd serverless-manager/
    ```

2. Provision the k3d cluster.

    ```bash
    kyma provision k3d
    ```

3. Install prerequisites.

    ```bash
    kyma deploy -s main --component cluster-essentials --profile production --ci
    ```

    > **NOTE:** This step is required only because `serverless-manager` is in the early stage, and there is no manager for the Kyma CRDs installation.

4. Build and push the Serverless Manager image.

    ```bash
    make module-image IMG_REGISTRY=localhost:5001/unsigned/operator-images IMG=localhost:5001/serverless-manager-dev-local:0.0.1
    ```

5. Build and push the Serverless module.

    ```bash
    make module-build IMG=k3d-kyma-registry:5001/serverless-manager-dev-local:0.0.1 MODULE_REGISTRY=localhost:5001/unsigned
    ```

6. Verify if the module and the manager's image are pushed to the local registry.

    ```bash
    curl localhost:5001/v2/_catalog
    ```

    You should get a result similar to this example:

    ```json
    {"repositories":["serverless-manager-dev-local","unsigned/component-descriptors/kyma.project.io/module/serverless"]}
    ```

7. Inspect the generated module template.

    > **NOTE:** The following sub-steps are temporary workarounds.

    Edit `template.yaml` and:

    - change `target` to `control-plane`

    ```yaml
    spec:
        target: control-plane
    ```

    > **NOTE:** This is required in the single cluster mode only.

    - change the existing repository context in `spec.descriptor.component`:  
    
    ```yaml
    repositoryContexts:                                                                           
      - baseUrl: k3d-kyma-registry.localhost:5000/unsigned
        componentNameMapping: urlPath                                                               
        type: ociRegistry
    ```

    > **NOTE:** Because Pods inside the k3d cluster use the docker-internal port of the registry, it tries to resolve the registry against port 5000 instead of 5001. K3d has registry aliases, but `module-manager` is not part of k3d and thus does not know how to properly alias `k3d-kyma-registry.localhost:5001`.

8. Install modular Kyma on the k3d cluster.

    This installs the latest versions of `module-manager` and `lifecycle-manager`.

    Use the `--template` flag to deploy the Serverless module manifest from the beginning, or apply it using kubectl later.

    ```bash
    kyma alpha deploy --template=./template.yaml
    ```

    Kyma installation is ready, but the module is not yet activated.

    ```bash
    kubectl get kymas.operator.kyma-project.io -A
    ```

    You should get a result similar to the following example:

    ```text
    NAMESPACE    NAME           STATE   AGE
    kcp-system   default-kyma   Ready   71s
    ```

    Serverless Module is a known module, but not activated.

    ```bash
    kubectl get moduletemplates.operator.kyma-project.io -A 
    ```

    You should get a result similar to the following example:

    ```text
    NAMESPACE    NAME                  AGE
    kcp-system   moduletemplate-serverless   2m24s
    ```

9. Give Module Manager permission to install CustomResourceDefinition (CRD) cluster-wide.

    `module-manager` must be able to apply CRDs to install modules. In the remote mode (with control-plane managing remote clusters) it gets an administrative kubeconfig, targeting the remote cluster to do so. In the local mode (single-cluster mode), it uses Service Account and does not have permission to create CRDs by default.

    Run the following command to make sure the module manager's Service Account gets an administrative role:

    ```bash
    kubectl edit clusterrole module-manager-manager-role
    ```

    And add the following element under `rules`:

    ```yaml
    - apiGroups:
      - "*"
      resources:
      - "*"                  
      verbs:                  
      - "*"
    ```

    > **NOTE:** This is a temporary workaround and is only required in the single-cluster mode.

10. Enable Serverless in the Kyma custom resource (CR)

    ```bash
    kubectl edit kymas.operator.kyma-project.io -n kcp-system default-kyma
    ```

    And add the following field under `spec`:

    ```yaml
      modules:
      - name: serverless
        channel: alpha
    ```

## Manual installation

1. Clone the project.

    ```bash
    git clone https://github.com/kyma-project/serverless-manager.git && cd serverless-manager/
    ```

2. Set the `serverless-manager` image name.

    ```bash
    export IMG=<DOCKER_USERNAME>/custom-serverless-manager:0.0.1
    ```

3. Verify the compability.

    ```bash
    make test
    ```

4. Build and push the image to the registry.

    ```bash
    make module-image
    ```

5. Deploy.

    ```bash
    make deploy
    ```

## Using `serverless-manager`

- Create a Serverless instance.

    ```bash
    kubectl apply -f config/samples/operator_v1alpha1_serverless_k3d.yaml
    ```

- Delete a Serverless instance.

    ```bash
    kubectl delete -f config/samples/operator_v1alpha1_serverless_k3d.yaml
    ```

- Update the Serverless properties.

    The following example shows how you can modify the Serverless docker registry address using the `serverless.operator.kyma-project.io` CR:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: operator.kyma-project.io/v1alpha1
    kind: Serverless
    metadata:
    name: serverless-sample
    spec:
    dockerRegistry:
        enableInternal: false
        registryAddress: k3d-kyma-registry:5000
        serverAddress: k3d-kyma-registry:5000
    EOF
    ```

## Troubleshooting

- For MacBook M1 users

    Some parts of the scripts may not work because the Kyma CLI is not released for Apple Silicon users. To fix it [install Kyma CLI manually](https://github.com/kyma-project/cli#installation) and export the path to it.

    ```bash
    export KYMA=$(which kyma)
    ```

    > The example error may look like this: `Error: unsupported platform OS_TYPE: Darwin, OS_ARCH: arm64; to mitigate this problem set variable KYMA with the absolute path to kyma-cli binary compatible with your operating system and architecture.  Stop.`
