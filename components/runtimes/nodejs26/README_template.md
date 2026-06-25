# Ejected Functions Workspace

This folder contains the files you need to test, run, deploy, and productize your business application:

* `handler.js` — source code of your business application
* `server.mjs` — server that loads and runs `handler.js`
* `package.json` — dependencies for `handler.js` and `server.mjs`
* `lib/` — internal server modules (tracing, metrics, request timeout, and graceful shutdown)
* `sdk/` — user-facing SDK module (`require('sdk')`) exposing CloudEvent helpers, tracer, and function metadata getters
* `resources/` — basic Kubernetes resources for deploying the application on a cluster
* `Dockerfile` — application image definition
* `Makefile` — automation scripts

## Scripts and Automations

Use the `Makefile` targets to run the application locally or build and deploy it on a k3d cluster. To see all available targets, run:

```bash
make help
```

### Run Application Locally

> [!NOTE]
> When you run the application outside the cluster, it cannot reach in-cluster services or read container environment variables. Use this target to test without such dependencies, or mock them and export the environment variables manually.

```bash
export FUNC_NAME=<name>
export FUNC_RUNTIME=<runtime>
export SERVICE_NAMESPACE=<namespace>

make run
```

### Deploy Application on k3d

To build, import, and deploy the application on a k3d cluster:

```bash
make k3d-deploy
```

### Build and Push Application

To build and push your image to the location specified by `IMG`:

```bash
make docker-build docker-push IMG=<image>
```
