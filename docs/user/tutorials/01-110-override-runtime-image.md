# Override Runtime Image

This tutorial shows how to build a custom runtime image and override the Function's base image with it.

## Prerequisites

Before you start, make sure you have these tools installed:

- [Serverless module installed](https://kyma-project.io/02-get-started/01-quick-install) in a cluster

## Steps

Follow these steps:

1. Follow [this example](https://github.com/kyma-project/serverless/tree/main/examples/custom-serverless-runtime-image) to build the Python's custom runtime image.


  > **NOTE:** Kyma Serverless enforces a strict Pod and container-level securityContext for all Functions (non-root execution, minimal Linux capabilities, and other hardening defaults). These constraints also apply for Functions with custom runtime image. Make sure that your custom image supports running as a non-root user under the restricted Pod security level (for example: no dependency on root ownership of writable paths, privileged operations, or added capabilities).

<!-- tabs:start -->

#### **Kyma CLI**

2. Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    export RUNTIME_IMAGE_URL={RUNTIME_IMAGE_URL} # image pull URL; for example {dockeruser}/foo:0.1.0
    ```

3. Create your local development workspace using the built image:

    ```bash
    mkdir {FOLDER_NAME}
    cd {FOLDER_NAME}
    kyma function init python
    ```

4. Deploy your Function:

    ```bash
    kyma function create python $NAME \
      --namespace $NAMESPACE --runtime python312 \
      --runtime-image-override $RUNTIME_IMAGE_URL \
      --source handler.py --dependencies requirements.txt
    ```

5. Verify whether your Function is running:

    ```bash
    kyma function get $NAME --namespace $NAMESPACE
    ```

#### **kubectl**

2. Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    export RUNTIME_IMAGE_URL={RUNTIME_IMAGE_URL} # image pull URL; for example {dockeruser}/foo:0.1.0
    ```

3. Create a Function CR that specifies the Function's logic:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: serverless.kyma-project.io/v1alpha2
   kind: Function
   metadata:
     name: $NAME
     namespace: $NAMESPACE
   spec:
     runtime: python312
     runtimeImageOverride: $RUNTIME_IMAGE_URL
     source:
       inline:
         source: |
           def main(event, context):
             return "hello world"
   EOF
   ```

4. Verify whether your Function is running:

    ```bash
    kubectl get functions $NAME -n $NAMESPACE
    ```

<!-- tabs:end -->
