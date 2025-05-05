# Overview

This example demostrates a Serverless Function that reads mounted service binding using [`@sap/xsenv`](https://www.npmjs.com/package/@sap/xsenv) in order to communicate with SAP BTP service instances. In this specific case, a Function is bound to a SAP BTP Objest Store service instance to read the objects from the store.

> **NOTE:** Please note, that resulting Storage Classes of Object Store Service are dependant on the underlying IaaS (see the [documentation](https://help.sap.com/docs/object-store/object-store-service-on-sap-btp/service-plans-and-entitlements-26c3918cae3049a7bb3aaa3c0b4edb55?version=Cloud&locale=en-US)). This example was written with the assumption that the SAP BTP subaccount is created on top of AWS IaaS and therefore  `@aws-sdk/client-s3` library is used in the Function code.


## Prerequisites

- Kyma cluster with [Serverless, API-Gateway and BTP-Operator modules installed](https://kyma-project.io/#/02-get-started/01-quick-install)
- Entitlement to [Object Store](https://help.sap.com/docs/object-store?locale=en-US) service usage on SAP BTP Platform (S3 compliant plan)


### Prerequisistes for Shared Service Instance Variant 
This example comes with a variant (commented out by default) that is based on the service instance sharing feature of the SAP BTP Platform. 
In that variant, an actual shared instance of Object Store service is needed together with the credentials for a Service Manager service binding belonging to the same BTP Subaccount where the `objectstore` instance is deployed.

To follow the variant with shared service instance: 
 1. Uncomment the secret generation part in the `kustomization.yaml` file an replace the `object-store-service-instance.yaml` with `object-store-service-reference-instance.yaml` in the resource listing. 
 2. Create a new `k8s-resources/service-manager-credentials.env` file from `k8s-resources/service-manager-credentials-template.env` with the values from Service Manager service binding.

If you want to explore this variant, read more on how to map shared service instances [here](https://help.sap.com/docs/btp/sap-business-technology-platform/namespace-level-mapping?locale=en-US&version=Cloud).

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

