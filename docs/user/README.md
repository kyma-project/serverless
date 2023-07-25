# Serverless module

## What is serverless?

"Serverless" refers to an architecture in which the infrastructure of your applications is managed by cloud providers. Contrary to its name, a serverless application does require a server but it doesn't require you to run and manage it on your own. Instead, you subscribe to a given cloud provider, such as AWS, Azure, or GCP, and pay a subscription fee only for the resources you actually use. Because the resource allocation can be dynamic and depends on your current needs, the serverless model is particularly cost-effective when you want to implement a certain logic that is triggered on demand. Simply, you get your things done and don't pay for the infrastructure that stays idle.

Kyma offers a service (known as "functions-as-a-service" or "FaaS") that provides a platform on which you can build, run, and manage serverless applications in Kubernetes. These applications are called **Functions** and they are based on the[Function custom resource (CR)](/docs/user/06-10-function.md) objects. They contain simple code snippets that implement a specific business logic. For example, you can define that you want to use a Function as a proxy that saves all incoming event details to an external database.

Such a Function can be:

- Triggered by other workloads in the cluster (in-cluster events) or business events coming from external sources. You can subscribe to them using a [Subscription CR](https://github.com/kyma-project/kyma/blob/main/docs/05-technical-reference/00-custom-resources/evnt-01-subscription.md).
- Exposed to an external endpoint (HTTPS). With an [APIRule CR](https://github.com/kyma-project/kyma/blob/main/docs/05-technical-reference/00-custom-resources/apix-01-apirule.md), you can define who can reach the endpoint and what operations they can perform on it.

## What is Serverless in Kyma?

Serverless in Kyma is an area that:

- Ensures quick deployments following a Function approach
- Enables scaling independent of the core applications
- Gives a possibility to revert changes without causing production system downtime
- Supports the complete asynchronous programming model
- Offers loose coupling of Event providers and consumers
- Enables flexible application scalability and availability

Serverless in Kyma allows you to reduce the implementation and operation effort of an application to the absolute minimum. It provides a platform to run lightweight Functions in a cost-efficient and scalable way using JavaScript and Node.js. Serverless in Kyma relies on Kubernetes resources like [Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/), [Services](https://kubernetes.io/docs/concepts/services-networking/service/) and [HorizontalPodAutoscalers](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) for deploying and managing Functions and [Kubernetes Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/) for creating Docker images.

## Serverless module

The Serverless module allows you to install, uninstall and configure Kyma's Serverless on your Kubernetes cluster, using Serverless Operator.

## Serverless Operator

When you enable the Serverless module, Serverless Operator takes care of installation and configuration of Serverless on your cluster. It manages the Serverless lifecycle based on the dedicated Serverless custom resource (CR).

## Useful links

If you're interested in learning more about the Serverless area, follow these links to:

- Perform some simple and more advanced tasks:

  - Create an [inline](/docs/user/03-10-create-inline-function.md) or a [Git](/docs/user/03-11-create-git-function.md) Function
  - [Expose the Function](/docs/user/03-20-expose-function.md)
  - [Manage Functions through Kyma CLI](/docs/user/03-30-manage-functions-with-kyma-cli.md)
  - [Debug a Function](/docs/user/03-40-debug-function.md)
  - [Synchronize Functions in a GitOps fashion](/docs/user/03-50-sync-function-with-gitops.md)
  - [Set an external Docker registry](/docs/user/03-60-set-external-registry.md) for your Function images and [switch between registries at runtime](/docs/user/03-70-switch-to-external-registry.md)
  - [Log into a private package registry](/docs/user/03-80-log-into-private-packages-registry.md)
  - [Set asynchronous communication between Functions](/docs/user/03-90-set-asynchronous-connection)
  - [Customize Function traces](/docs/user/03-100-customize-function-traces.md)
  - [Override runtime image](/docs/user/03-110-override-runtime-image.md)
  - [Inject environment variables](/docs/user/03-120-inject-envs.md)
  - [Use external scalers](/docs/user/03-130-use-external-scalers.md)
  - [Access to Secrets mounted as Volume](/docs/user/03-140-use-secret-mounts.md)

- Troubleshoot Serverless-related issues when:

   - [Functions won't build](/docs/user/03-10-cannot-build-functions.md)
   - [Container fails](/docs/user/03-20-failing-function-container.md)
   - [Debugger stops](/docs/user/03-30-function-debugger-strange-location.md)

- Analyze Function specification and configuration files:

  - [Function](/docs/user/06-10-function.md) custom resource
  - [`config.yaml` file](/docs/user/07-60-function-configuration-file.md) in Kyma CLI
  - [Function specification details](/docs/user/07-70-function-specification.md)

- Understand technicalities behind Serverless implementation:

  - [Serverless architecture](/docs/user/04-10-architecture.md) and [Function processing](/docs/user/07-20-function-processing-stages.md)
  - [Switching registries](/docs/user/07-30-switching-registries.md)
  - [Git source type](/docs/user/07-40-git-source-type.md)
  - [Exposing Functions](/docs/user/07-50-exposing-functions.md)
  - [Available presets](/docs/user/07-80-available-presets.md)
  - [Environment variables in Functions](/docs/user/05-20-env-variables.md)