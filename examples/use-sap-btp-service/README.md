# Overview
<!-- markdown-link-check-disable-next-line -->
This example demonstrates a Serverless Function that reads mounted service binding using [`@sap/xsenv`](https://www.npmjs.com/package/@sap/xsenv) to communicate with SAP BTP service instances. In this specific case, a Function is bound to an SAP BTP Object Store service instance to read objects from the store.

> [!NOTE]
> Resulting Storage Classes of Object Store Service are dependent on the underlying IaaS (see [Service Plans and Entitlements](https://help.sap.com/docs/object-store/object-store-service-on-sap-btp/service-plans-and-entitlements-26c3918cae3049a7bb3aaa3c0b4edb55?version=Cloud&locale=en-US)). This example assumes that the SAP BTP subaccount is created on top of AWS IaaS. Therefore, the  `@aws-sdk/client-s3` library is used in the Function code.


## Prerequisites

- Kyma cluster with the [Serverless, API Gateway, and SAP BTP Operator modules installed](https://kyma-project.io/#/02-get-started/01-quick-install)
- Entitlement to the [Object Store](https://help.sap.com/docs/object-store?locale=en-US) service usage on SAP BTP (S3 compliant plan)


### Prerequisites for Shared Service Instance Variant 
This example includes a variant (commented out by default) based on SAP BTP's service instance sharing feature. 
It requires an actual shared instance of the Object Store service together with the credentials for a Service Manager service binding belonging to the same SAP BTP Subaccount where the `objectstore` instance is deployed.

To use the variant with the shared service instance, follow these steps: 
 1. Uncomment the Secret generation part in the `kustomization.yaml` file and replace `object-store-service-instance.yaml` with `object-store-service-reference-instance.yaml` in the resource listing. 
 2. Create a new `k8s-resources/service-manager-credentials.env` file from `k8s-resources/service-manager-credentials-template.env` with the values from the Service Manager service binding.

If you want to explore this variant, read more on [how to map shared service instances](https://help.sap.com/docs/btp/sap-business-technology-platform/namespace-level-mapping?locale=en-US&version=Cloud).

## Deploy 

To deploy the example use:

   ```bash
   make deploy
   ```

## Test

To list objects in the bucket:

   ```bash
   export DOMAIN={use your cluster domain}
   curl https://object-store.$DOMAIN    
   ```

