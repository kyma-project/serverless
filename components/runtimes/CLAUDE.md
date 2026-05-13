# Runtimes

Pre-built Docker images that host user-defined serverless functions. Each runtime is a small HTTP server that loads and executes user-supplied handler code at container startup.

## Architecture

There are two generations of handler API:

### Legacy API (nodejs20, nodejs22, nodejs24, python312)
- Handler signature: `main(event, context)` 
- `event` wraps the HTTP request with CloudEvent parsing, tracer, and response helpers
- `context` provides function metadata (name, namespace, timeout, runtime)
- Python312 uses Bottle + CherryPy; Node.js uses Express

### New API (nodejs26, python314)
- **Node.js 26**: Handler receives raw Express `(req, res)` objects directly
- **Python 314**: Handler takes no arguments; uses Flask's `flask.request` context-local
- Shared `sdk` module provides tracing, CloudEvent helpers, and function metadata
- No `event`/`context` wrapper — direct framework access

## Common Structure

Each runtime directory contains:
- `server.mjs` / `server.py` — Main entry point, HTTP server setup, middleware
- `lib/` — Internal modules (tracer, metrics, helper)
- `sdk/` (nodejs26) or `lib/sdk.py` (python314) — User-facing SDK module
- `Dockerfile` — Multi-stage build with optional FIPS variant
- `Makefile` — `docker-build`, `docker-push`, `k3d-deploy`, `run`

## Running Locally

```bash
make -C components/runtimes/nodejs26 run    # npm install + npm start
make -C components/runtimes/python314 run   # venv + pip install + python server.py
```

Note: without a `handler.js`/`handler.py` in the expected path, the server starts but logs a "function not loaded" error. Create a test handler to exercise locally.

## Key Environment Variables

All runtimes read configuration from env vars injected by the Function Controller:

| Variable | Description |
|----------|-------------|
| `FUNC_NAME` | Function name |
| `SERVICE_NAMESPACE` | Kubernetes namespace |
| `FUNC_RUNTIME` | Runtime identifier (e.g. `nodejs26`, `python314`) |
| `FUNC_TIMEOUT` | Request timeout in seconds (default: 180) |
| `FUNC_HANDLER` | Exported function name (default: `main`) |
| `MOD_NAME` | Handler module filename without extension (default: `handler`) |
| `FUNCTION_PATH` | Path to user source code (default: `/kubeless`) |
| `TRACE_COLLECTOR_ENDPOINT` | OTLP trace collector URL |
| `PUBLISHER_PROXY_ADDRESS` | Eventing publisher proxy URL |
| `REQ_MB_LIMIT` | (Node.js) Body size limit in MB |
| `FUNC_MEMFILE_MAX` | (Python) Body size limit in bytes |
| `SERVER_NUMTHREADS` | (Python) gevent spawn pool size |
| `KYMA_INTERNAL_LOGGER_ENABLED` | Enable Apache combined request logging |

## SDK Module

Both new runtimes expose an `sdk` module for user handlers:

- **Node.js 26**: `import { getCloudEvent, emitCloudEvent, getTracer } from 'sdk'`
- **Python 314**: `import sdk` then `sdk.get_cloud_event()`, `sdk.emit_cloud_event()`, `sdk.get_tracer()`

The SDK is configured once at server startup via `configure()`/`_configure()` — user handlers get the pre-configured singleton.

## Testing Changes

There are no unit tests for runtimes currently. To verify changes:
1. `make run` in the runtime directory with a test handler file
2. `curl http://localhost:8080/` to invoke the function
3. `curl http://localhost:8080/healthz` for health check
4. `curl http://localhost:8080/metrics` for Prometheus metrics
