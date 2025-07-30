# Buildless Mode of Serverless

## Overview: Moving Towards Buildless Mode

From the beginning, Kyma Serverless has aimed to accelerate the development of fast prototypes by allowing users to focus on business logic rather than containerization and Kubernetes deployment. Our goal is to remove operational barriers so developers can iterate quickly and efficiently.

With the introduction of buildless mode, we are taking this vision further. We significantly shorten the feedback loop during prototype development by eliminating the image build step in Kyma runtime. In buildless mode, instead of building and pushing custom function images into the in-cluster registry, your code and dependencies are mounted into Kyma-provided runtime images. This approach positions Kyma Serverless as a more efficient development tool, enabling even faster iteration. Additionally, it eliminates the architectural complexities and limitations of deploying Serverless Functions on Kubernetes.

## Benefits

- **Faster deployment**: Functions deploy faster as the build job is no longer required.
- **Resource efficiency**: Eliminates the need for Serverless to acquire computational resources from worker nodes to build the image.
- **Enhanced security**: By eliminating build jobs, Functions can run in namespaces with more restrictive Pod security levels enabled.
- **No additional storage required**: No additional storage resources are used to store the Function image.
- **Simplified Architecture**: The Serverless module no longer requires Docker Registry, making it more lightweight and easier to manage.

## Changes When Switching to Buildless Serverless

- The internal resources used for storing custom Function images (Docker Registry) are uninstalled from the Serverless module
- Your Functions start quicker as build Jobs for Functions are no longer created, and the existing Job resources are deleted.
- Libraries and dependencies are downloaded dynamically at the start of each Function's execution.
- Function code is directly injected into the runtime Pod, removing the need for pre-built images.
- Your existing Functions are redeployed without downtime and started as Pods based on Kyma-provided images with your code and dependencies mounted.

## Using Fixed Dependency Versions

- **Avoid using `latest` versions of Function dependencies**: Since dependencies are resolved at Function's Pod start time in buildless mode, using `latest` versions can lead to inconsistencies between replicas of the same Function. This may be the case when the dependency provider releases a new version after one replica is already running and before another replica is created due to auto-scaling.  Always specify exact versions of dependencies to ensure stability and predictability.
- **Dependency resolution behavior**: Be aware that each replica of a Function may resolve and use a different dependency version if the version is not explicitly pinned.

## Switching to Buildless Serverless

To enable buildless mode for Serverless, you must enable it in the Serverless custom resource (CR) annotations. Follow these steps:

1. Edit the Serverless CR:
   ```yaml
   kubectl edit -n kyma-system serverlesses.operator.kyma-project.io default
   ```
   
2. In the metadata section of the CR, add the following annotation:
   ```yaml
    annotations:
      serverless.kyma-project.io/buildless-mode: "enabled"
   ```

3. Save the file to apply the changes.