# Custom Serverless Runtime Image

## Overview

This example shows how to create own custom runtime for a Serverless Function based on the Python runtime and the `debian:bullseye-slim` base image to provide support for glibc.

## Prerequisites

- Docker as a build tool

## Build an Example Runtime

1. Export the following environments:

   ```bash
   export IMAGE_NAME=<image_name>
   export IMAGE_TAG=<image_tag>
   ```

2. Build and push the image:

   ```bash
   docker build -t "${IMAGE_NAME}/${IMAGE_TAG}" .
   docker push "${IMAGE_NAME}/${IMAGE_TAG}"
   ```

> [!NOTE]
> You can use it to define your Functions in Kyma. To learn more, read [how to override runtime image](https://kyma-project.io/#/serverless-manager/user/resources/06-20-serverless-cr?id=custom-resource-parameters).
