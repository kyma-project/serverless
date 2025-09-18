# Ejected Functions Workspace

This folder contains all most important block to test, run, deploy and productize your buisness application and consists of following elements:

* `handler.js` - source code of the buisness application
* `server.mjs` - server source code needed to run the handler.js
* `package.json` - dependencies file required for the handler.js and the server.mjs to run
* `lib/` - directory containing server SDK (like cloudevents or tracing functionality)
* `resources/` - directory with basic Kubernetes resources needed to deploy application on a cluster
* `Dockerfile` - application image definition
* `Makefile` - basic portion of automations and scripts

## Scripts and Automations

The `Makefile` file is designed to speed up some processes like running application locally, build or deploy it on a k3d cluster.

Read more about possibilities and functionalites by running the `make help` target.

### To run application locally

>NOTE: because application is run outside of the cluster it cannot simply reach in-cluster services and use container envs. It is strongly recommend to use following target to test application without such dependecies or mock them and export container envs manually.

```bash
export FUNC_NAME=<name>
export FUNC_RUNTIME=<runtime>
export SERVICE_NAMESPACE=<SERVICE_NAMESPACE>

make run
```

### To deploy application on k3d

The workspace is designed to easly start working on productization of buisness application ejected from the Function and allows to test it on k3d cluster by building it, importing and deploying on it:

```bash
make k3d-deploy
```

### To build and push application

Build and push your image to the location specified by `IMG`:

```bash
make docker-build docker-push IMG=<image>
```
