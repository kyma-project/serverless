# Serverless Architecture Updates

## Changes

- The internal Docker Registry is no longer part of the Serverless module. Instead, the Docker Registry is now a separate, standalone module.
- There is no longer a build job for Functions. Instead, a base image is used, which mounts the required dependencies dynamically.
- Libraries and dependencies are downloaded at the start of the Function's execution. This means that each replica of the Function can potentially use a different version of the dependencies.
- Function code is now injected directly into the runtime Pod, eliminating the need for pre-built images.

## Benefits

- **Simplified architecture**: By separating the Docker Registry into its own module, the Serverless module is now more lightweight and easier to manage.
- **Faster deployment**: The removal of the build job reduces the time required to deploy Functions.
- **Dynamic dependency resolution**: Dependencies are resolved at runtime, allowing for more flexibility in managing library versions.
- **Improved flexibility**: Injecting Function code into the runtime Pod simplifies the deployment process and reduces image management overhead.

## Managing Dependencies

- **Avoid using `latest` versions of Function dependencies**: Since dependencies are resolved at Function's Pod start time runtime in build-less mode, using `latest` versions can lead to inconsistencies between replicas of the same Function. This may be the case when dependency provider releases new version after one replica is already running and before another replica was created due to auto-scaling.  Always specify exact versions of dependencies to ensure stability and predictability.
- **Dependency resolution behavior**: Be aware that each replica of a Function may resolve and use a different version of a dependency if the version is not explicitly pinned.