# Expose a Function Using the APIRule Custom Resource

This tutorial shows how you can expose your Function to access it outside the cluster, through an HTTP proxy. To expose it, use an [APIRule custom resource (CR)](https://kyma-project.io/docs/kyma/latest/05-technical-reference/00-custom-resources/apix-01-apirule/). Function Controller reacts to an instance of the APIRule CR and, based on its details, it creates an Istio VirtualService and Oathkeeper Access Rules that specify your permissions for the exposed Function.

When you complete this tutorial, you get a Function that:

- Uses the `no_auth` handler, allowing access on an unsecured endpoint.
- Accepts the `GET`, `POST`, `PUT`, and `DELETE` methods.

To learn more about securing your Function, see the [Expose and secure a workload with OAuth2](https://kyma-project.io/docs/kyma/latest/03-tutorials/00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-01-expose-and-secure-workload-oauth2/) or [Expose and secure a workload with JWT](https://kyma-project.io/docs/kyma/latest/03-tutorials/00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-03-expose-and-secure-workload-jwt/) tutorials.

Read also about [Function’s specification](../technical-reference/07-70-function-specification.md) if you are interested in its signature, `event` and `context` objects, and custom HTTP responses the Function returns.

## Prerequisites

- You have an [existing Function](01-10-create-inline-function.md).
- You have the [Istio, API Gateway and Serverless modules added](https://kyma-project.io/#/02-get-started/01-quick-install).
- For the Kyma CLI scenario, you have Kyma CLI installed.

## Procedure

You can expose a Function using Kyma dashboard, Kyma CLI, or kubectl:

<!-- tabs:start -->

#### **Kyma Dashboard**

1. Select a namespace from the drop-down list in the navigation panel. Make sure the namespace includes the Function that you want to expose using the APIRule CR.

2. Go to **Discovery and Network** > **API Rules**, and choose **Create**.

3. Enter the following information:

    - The APIRule's **Name** matching the Function's name.

    > [!NOTE]
    > The APIRule CR can have a name different from that of the Function, but it is recommended that all related resources share a common name.

    - **Service Name** matching the Function's name.

    - **Host** to determine the host on which you want to expose your Function.

4. Edit the access strategy in the **Rules** > **Access Strategies** section
  - Select the methods `GET`, `POST`, `PUT`, and `DELETE`. 
  - Use the default `no_auth` handler.

5. Select **Create** to confirm your changes.

6. Check if you can access the Function by selecting the HTTPS link under the **Host** column for the newly created APIRule. If successful, the `Hello World!` message appears.

#### **Kyma CLI**

1. Run the following command to get the domain name of your Kyma cluster:

    ```bash
    kubectl get gateway -n kyma-system kyma-gateway \
        -o jsonpath='{.spec.servers[0].hosts[0]}'
    ```

2. Export the result without the leading `*.` as an environment variable:

    ```bash
    export DOMAIN={DOMAIN_NAME}

3. Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={NAMESPACE_NAME}
    export KUBECONFIG={PATH_TO_YOUR_KUBECONFIG}
    ```

4. Download the latest configuration of the Function from the cluster. This way, you update the local `config.yaml` file with the Function's code.

    ```bash
    kyma sync function $NAME -n $NAMESPACE
    ```

5. Edit the local `config.yaml` file and add the **apiRules** schema for the Function at the end of the file:

    ```yaml
    apiRules:
        - name: {FUNCTION_NAME}
          service:
            host: {FUNCTION_NAME}.{DOMAIN_NAME}
          rules:
            - methods:
                - GET
                - POST
                - PUT
                - DELETE
              accessStrategies:
                - handler: no_auth
    ```

6. Apply the new configuration to the cluster:

    ```bash
    kyma apply function
    ```

7. Check if the Function's code was pushed to the cluster and reflects the local configuration:

    ```bash
    kubectl get apirules $NAME -n $NAMESPACE
    ```

If successful, the APIRule has the status `OK`.

    ```bash
    kubectl get apirules $NAME -n $NAMESPACE -o=jsonpath='{.status.APIRuleStatus.code}'
    ```

1. Call the Function's external address:

    ```bash
    curl https://$NAME.$DOMAIN
    ```

    If successful, the `Hello World!` message appears.

#### **kubectl**

1. Run the following command to get the domain name of your Kyma cluster:

    ```bash
    kubectl get gateway -n kyma-system kyma-gateway \
        -o jsonpath='{.spec.servers[0].hosts[0]}'
    ```

2. Export the result without the leading `*.` as an environment variable:

    ```bash
    export DOMAIN={DOMAIN_NAME}

3. Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    export KUBECONFIG={PATH_TO_YOUR_KUBECONFIG}
    ```

    > [!NOTE]
    > The APIRule CR can have a name different from that of the Function, but it is recommended that all related resources share a common name.

4. Create an APIRule CR, which exposes your Function on port `80`.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      gateway: kyma-system/kyma-gateway
      host: $NAME.$DOMAIN
      service:
        name: $NAME
        port: 80
      rules:
        - path: /.*
          methods:
            - GET
            - POST
            - PUT
            - DELETE
          accessStrategies:
            - handler: no_auth
    EOF
    ```

5. Check that the APIRule was created successfully and has the status `OK`:

    ```bash
    kubectl get apirules $NAME -n $NAMESPACE -o=jsonpath='{.status.APIRuleStatus.code}'
    ```

6. Access the Function's external address:

    ```bash
    curl https://$NAME.$DOMAIN
    ```

    If successful, the `Hello World!` message appears.

<!-- tabs:end -->