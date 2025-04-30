# Overview

This example demostrates a Serverless Function that reads mounted service binding in order to communicate with SAP BTP service instances. Here, SAP BTP Objest Store service instance is used as an example but any other service instance from SAP BTP platform can be bound the same way using [`@sap/xsenv`](https://www.npmjs.com/package/@sap/xsenv) package.

## Prerequisites

- [Serverless and API-Gateway modules installed](https://kyma-project.io/#/02-get-started/01-quick-install) locally or on a cluster
- Shared instance of `objectstore` service on SAP BTP Platform 
- Credentials pointing to the Service Manager service binding in the same subaccount where the `objectstore` instance is deployed (see [documentation](https://help.sap.com/docs/btp/sap-business-technology-platform/namespace-level-mapping?locale=en-US&version=Cloud)).

## Deploy 

Create a new `k8s-resources/service-manager-credentials.env` file from `k8s-resources/service-manager-credentials-template.env`. Provide the values from Service Manager service binding.

   ```bash
   make deploy
   ```

## Test

To list objects in the bucket:

   ```bash
   export DOMAIN={use your cluster domain}
   curl https://object-store.$DOMAIN    
   ```

