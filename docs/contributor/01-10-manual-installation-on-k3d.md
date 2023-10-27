## Manual installation on the k3d cluster

1. Clone the project.

    ```bash
    git clone https://github.com/kyma-project/serverless-manager.git && cd serverless-manager/
    ```

2. Provision the k3d cluster.

    ```bash
    kyma provision k3d
    ```

3. Build and push the Serverless Operator image.

    ```bash
    make module-image IMG_REGISTRY=localhost:5001/unsigned/operator-images IMG=localhost:5001/serverless-operator-dev-local:0.0.1
    ```

4. Build and push the Serverless module.

    ```bash
    make module-build IMG=k3d-kyma-registry:5001/serverless-operator-dev-local:0.0.1 MODULE_REGISTRY=localhost:5001/unsigned
    ```

5. Verify if the module and the operator's image are pushed to the local registry.

    ```bash
    curl localhost:5001/v2/_catalog
    ```

    You should get a result similar to this example:

    ```json
    {
        "repositories": [
            "serverless-operator-dev-local",
            "unsigned/component-descriptors/kyma-project.io/module/serverless"
        ]
    }
    ```

6. Inspect the generated module template.

    > **NOTE:** The following sub-steps are temporary workarounds.

    Edit `moduletemplate.yaml` under the `config/moduletemplates` folder and:

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

    > **NOTE:** Because Pods inside the k3d cluster use the docker-internal port of the registry, it tries to resolve the registry against port 5000 instead of 5001. K3d has registry aliases, but `lifecycle-manager` is not part of k3d and thus does not know how to properly alias `k3d-kyma-registry.localhost:5001`.

7. Install modular Kyma on the k3d cluster.

    This installs the latest versions of `lifecycle-manager`.

    Use the `--template` flag to deploy the Serverless module manifest from the beginning, or apply it using kubectl later.

    ```bash
    kyma alpha deploy --templates=./config/moduletemplates/moduletemplate.yaml
    ```

    Kyma installation is ready, but the module is not yet activated.

    ```bash
    kubectl get kymas.operator.kyma-project.io -A
    ```

    You should get a result similar to the following example:

    ```bash
    NAMESPACE    NAME           STATE   AGE
    kyma-system   default-kyma   Ready   71s
    ```

    Serverless Module is a known module, but not activated.

    ```bash
    kubectl get moduletemplates.operator.kyma-project.io -A 
    ```

    You should get a result similar to the following example:

    ```bash
    NAMESPACE    NAME                  AGE
    kcp-system   moduletemplate-serverless   2m24s
    ```

8. Give Lifecycle Manager permission to install CustomResourceDefinition (CRD) cluster-wide.

    `lifecycle-manager` must be able to apply CRDs to install modules. In the remote mode (with control-plane managing remote clusters) it gets an administrative kubeconfig, targeting the remote cluster to do so. In the local mode (single-cluster mode), it uses Service Account and does not have permission to create CRDs by default.

    Run the following command to make sure the Lifecycle Manager's Service Account gets an administrative role:

    ```bash
    kubectl edit clusterrole lifecycle-manager-manager-role
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

9. Enable Serverless in the Kyma custom resource (CR)

    ```bash
    kyma alpha enable module serverless -c fast
    ```