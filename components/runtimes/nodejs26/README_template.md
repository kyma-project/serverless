# Ejected Functions Workspace

This folder contains the most important block to test, run, deploy, and productize your business application and consists of the following elements:

* `handler.js` - source code of the buisness application
* `server.mjs` - server source code required to run `handler.js`
* `package.json` - dependencies file required for the `handler.js` and the `server.mjs` files to run
* `lib/` - directory containing server SDK (like cloudevents or tracing functionality)
* `resources/` - directory with basic Kubernetes resources required to deploy the application on a cluster
* `Dockerfile` - application image definition
* `Makefile` - basic portion of automations and scripts

## Scripts and Automations

The `Makefile` file is designed to speed up processes such as running an application locally or building or deploying it on a k3d cluster.

Read more about possibilities and functionalites by running the `make help` target.

### Run Application Locally

> [!NOTE] 
> Because the application is run outside the cluster, it cannot simply reach in-cluster services and use container envs. It is strongly recommended to use the following target to test the application without such dependencies or mock them and export container envs manually.

```bash
export FUNC_NAME=<name>
export FUNC_RUNTIME=<runtime>
export SERVICE_NAMESPACE=<SERVICE_NAMESPACE>

make run
```

### Deploy Application on k3d

The workspace is designed to easily start working on the productization of a business application extracted from Function. It allows testing it on a k3d cluster by building, importing, and deploying it:

```bash
make k3d-deploy
```

### Build and Push Application

Build and push your image to the location specified by `IMG`:

```bash
make docker-build docker-push IMG=<image>
```
