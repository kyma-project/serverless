# Overview

This example provides a Serverless Function that exposes REST API to manage records stored in the redis as a backend. 

## Prerequisites

- [Serverless and API-Gateway modules installed](https://kyma-project.io/#/02-get-started/01-quick-install) locally or on a cluster

## Deploy 

1. To deploy this example run the `deploy` make target:

   ```bash
   make deploy
   ```

## Test

Send various HTTP requests to create, read and delete objects via the `rest-fn` Function:

   ```bash
   export DOMAIN={use your cluster domain}

   curl -H "Content-Type: application/json" -X POST  https://rest-fn.$DOMAIN/foo -d '{"value":"bar"}'

   curl https://rest-fn.$DOMAIN/foo 

   curl -X DELETE  https://rest-fn.$DOMAIN/foo

   ```

## Undeploy 

1. To undeploy this example run the `undeploy` make target:

   ```bash
   make undeploy
   ```