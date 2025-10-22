# Build-less mode of Serverless

## Benefits

- **Simplified architecture**: By separating the Docker Registry into its own module, the Serverless module is now more lightweight and easier to manage.
- **Faster deployment**: The removal of the build job reduces the time required to deploy Functions.
- **Dynamic dependency resolution**: Dependencies are resolved at runtime, allowing for more flexibility in managing library versions.
- **Improved flexibility**: Injecting Function code into the runtime Pod simplifies the deployment process and reduces image management overhead.
- **Reduced resource consumption**: Eliminating build jobs means Serverless no longer requires computational resources from worker nodes for image building.
- **Enhanced security**: By removing build jobs, Functions can run in namespaces with more restrictive Pod security levels enabled.

## What Changes with Buildless Mode

- Your Serverless module no longer includes an internal Docker Registry.
- Function builds are eliminated. Your Functions use a base image that mounts dependencies dynamically at runtime.
- Function dependencies are downloaded each time a Function Pod starts. This means different replicas of the same Function may use different dependency versions if you don't pin exact versions.
- Your Function code is injected directly into runtime Pods without requiring pre-built container images.

## Use Fixed Dependency Versions

- **Avoid using `latest` versions of Function dependencies**: Since dependencies are resolved at the Function's Pod start time in buildless mode, using `latest` versions can lead to inconsistencies between replicas of the same Function. This may be the case when the dependency provider releases a new version after one replica is already running and before another replica is created due to auto-scaling. Always specify exact versions of dependencies to ensure stability and predictability.
- **Dependency resolution behavior**: Be aware that each replica of a Function may resolve and use a different version of a dependency if the version is not explicitly pinned.