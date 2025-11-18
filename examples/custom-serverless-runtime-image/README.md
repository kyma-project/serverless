# Custom Serverless Runtime Image

## Overview

This example shows how to create own custom runtime for a Serverless Function based on the Python 3.12 runtime and the `python:3.12.3--slim-bookworm` base image to provide support for glibc.

## Prerequisites

- Docker as a build tool

## Build an Example Runtime

1. Export the following environments:

   ```bash
   export IMAGE_NAME=<image_name>
   export IMAGE_TAG=<image_tag>
   export REGISTRY_URL=<registry_url> # use your dockerhub username to use docker.io
   ```

2. Build and push the image:

   ```bash
   docker build  --platform linux/amd64 -t "${IMAGE_NAME}:${IMAGE_TAG}" .
   docker tag "${IMAGE_NAME}:${IMAGE_TAG}" "${REGISTRY_URL}/${IMAGE_NAME}:${IMAGE_TAG}"
   docker push "${REGISTRY_URL}/${IMAGE_NAME}:${IMAGE_TAG}"
   ```

> [!NOTE]
> You can use it to define your Functions in Kyma. To learn more, read [how to override runtime image](https://kyma-project.io/external-content/serverless/docs/user/resources/06-20-serverless-cr.html#custom-resource-parameters).
