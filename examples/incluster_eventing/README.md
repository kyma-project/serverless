# Asynchronous Communication Between Functions

## Overview

This example provides a simple scenario of asynchronous communication between two Functions, where:

- The first Function accepts the incoming traffic using HTTP, sanitizes the payload, and publishes the content as an in-cluster event by the [Kyma Eventing module](https://kyma-project.io/external-content/eventing-manager/docs/user/README).
- The second Function is a message receiver. It subscribes to the given event type and stores the payload.

This example also provides a template for a git project with Kyma Functions. Please refer to the Deploy section below.

## Prerequisites

- [Kyma CLI](https://github.com/kyma-project/cli)
- Kyma installed locally or on a cluster

## Deploy

Deploy to Kyma runtime using the `make deploy` target.

### Auto-Deploy Code Changes

Changes pushed to the `handler.js` files should be automatically pulled by Kyma Serverless as both Functions are of a Git type and reference this Git repository as the source.

### Test the Application

Send an HTTP request to the emitter Function.

   ```bash
   curl -H "Content-Type: application/json" -X POST  https://incoming.{your cluster domain} -d '{"foo":"bar"}'
   Event sent%
   ```

Fetch the logs of the receiver Function to observe the incoming message.

   ```bash
   > nodejs22-runtime@0.1.0 start
   > node server.mjs

   user code loaded in 0sec 0.649274ms
   storing data...
   {"foo":"bar"}
   ```
