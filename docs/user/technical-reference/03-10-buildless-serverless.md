# Build-less mode of Serverless

## Overview: Moving Towards Buildless Mode

From the beginning, Kyma Serverless has aimed to accelerate the development of fast prototypes by allowing users to focus on business logic rather than containerization and Kubernetes deployment. Our goal is to remove operational barriers so developers can iterate quickly and efficiently.

With the introduction of Buildless mode, we are taking this vision further. By eliminating the image build step in the Kyma runtime, we significantly shorten the feedback loop during prototype development. In buildldess mode, instead of building and pushing custom function images into in-cluster registry, your code and dependencies are simply mounted into Kyma-provided runtime images. This approach positions kyma serverless as more efficient development tool, enabling even faster iteration. Additionally it eliminates the architectural complexities and limitations of deploying serverless Functions on Kubernetes.

## Benefits

- **Simplified management**: The Serverless module is now more lightweight and easier to handle due to the separation of the Docker Registry into its own module.
- **Faster deployment**: Functions deploy faster as the build job is no longer required.
- **Dynamic dependency resolution**: Dependencies are resolved dynamically at runtime, offering greater adaptability in managing library versions.
- **Streamlined process**:  Function code is directly injected into the runtime Pod, reducing complexity and image management efforts.
- **Resource efficiency**: Eliminates the need to for serverless to acquire computational resources from worker nodes to build the image.
- **Enhanced security**: By eliminating build jobs functions can run in namespaces with more restrictive pod security levels enabled.

## What would change if I switch buildless on

- The internal resources used for storing custom function images (Docker Registry) will be uninstalled from the Serverless module
- Your  Functions will start quicker as build Jobs for Functions are no longer created (and existing Jobs resources will be removed).
- Libraries and dependencies are downloaded dynamically at the start of each Function's execution.
- Function code is directly injected into the runtime Pod, removing the need for pre-built images.

## Use fixed dependency versions

- **Avoid using `latest` versions of Function dependencies**: Since dependencies are resolved at Function's Pod start time runtime in build-less mode, using `latest` versions can lead to inconsistencies between replicas of the same Function. This may be the case when dependency provider releases new version after one replica is already running and before another replica was created due to auto-scaling.  Always specify exact versions of dependencies to ensure stability and predictability.
- **Dependency resolution behavior**: Be aware that each replica of a Function may resolve and use a different version of a dependency if the version is not explicitly pinned.