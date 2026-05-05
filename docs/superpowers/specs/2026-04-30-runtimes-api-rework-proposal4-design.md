# Serverless Runtimes API Rework — Proposal 4 Design Spec

**Issue:** [#2292](https://github.com/kyma-project/serverless/issues/2292)
**Date:** 2026-04-30
**Runtimes in scope:** `python314`, `nodejs26`

---

## Overview

Rework the serverless runtimes API (Proposal 4) to use native framework arguments:

- **Python**: no-args handler, Flask-native request access via `from flask import request`, Flask-native response via `return`
- **Node.js**: `(req, res)` passed directly to `main`, Express-native, return value ignored
- **SDK**: importable module bundled with each runtime, providing tracer, CloudEvent helpers, and function metadata that was previously spread across `event` and `context`

This is a breaking change — the old `(event, context)` signature is dropped with no compatibility shim.

---

## Handler API

### Python

```python
from flask import request
import sdk

def main():
    data = request.get_json()
    return {"runtime": sdk.get_runtime(), "name": sdk.get_function_name()}, 200
```

### Node.js

```javascript
import sdk from 'sdk';

export function main(req, res) {
    const ce = sdk.getCloudEvent();  // null if not a CloudEvent request
    res.json({ runtime: sdk.getRuntime(), body: req.body });
}
```

Node.js handler return values are ignored — Express contract applies: use `res` to respond.

---

## SDK Surface

Importable as `import sdk` (Python) / `import sdk from 'sdk'` (Node.js). Configured once at server startup via `sdk._configure(...)` before any request is served.

| Python | Node.js | Returns | Description |
|---|---|---|---|
| `get_tracer()` | `getTracer()` | OTel `Tracer` | Configured tracer — replaces `event.tracer` |
| `get_cloud_event()` | `getCloudEvent()` | CloudEvent or `None`/`null` | Parsed incoming CloudEvent — replaces CE fields on `event` |
| `emit_cloud_event(type, source, data, attrs?)` | `emitCloudEvent(type, source, data, attrs?)` | — | Publish outgoing CloudEvent — replaces `event.emitCloudEvent` |
| `get_function_name()` | `getFunctionName()` | `str` / `string` | Replaces `context['function-name']` |
| `get_namespace()` | `getNamespace()` | `str` / `string` | Replaces `context['namespace']` |
| `get_runtime()` | `getRuntime()` | `str` / `string` | Replaces `context['runtime']` |
| `get_timeout()` | `getTimeout()` | `float` / `number` | Replaces `context['timeout']` |

`get_request()` / `getRequest()` are intentionally omitted — Python users use Flask's `flask.request` directly, Node.js users use the `req` argument passed to `main`.

`memory-limit` from the old `context` is not carried over — it was already deprecated.

### SDK wiring

The server calls `sdk._configure(tracer, publisher_proxy_address, func_name, func_namespace, func_runtime, server_call_timeout)` once at startup. The SDK stores these as module-level variables. All configured values are server-level (not per-request), so no per-request isolation mechanism is needed in either runtime.

---

## ENV Architecture

All ENVs have defaults inside the runtime container. The controller injects only what it cannot calculate itself.

| ENV | Default in runtime | Injected by controller | Notes |
|---|---|---|---|
| `FUNC_NAME` | `''` | yes | unchanged |
| `FUNC_NAMESPACE` | `''` | yes | replaces `SERVICE_NAMESPACE` |
| `FUNC_RUNTIME` | `python314` / `nodejs26` | yes | unchanged |
| `SERVER_HOST` | `0.0.0.0` | no | |
| `SERVER_PORT` | `8080` | no | |
| `SERVER_NUMTHREADS` | `50` | no | Python only |
| `SERVER_CALL_TIMEOUT` | `180` | no | replaces `FUNC_TIMEOUT` |
| `HANDLER_FOLDER` | `/kubeless` | yes | Python only — replaces `FUNCTION_PATH` |
| `HANDLER_MODULE_NAME` | `handler` | yes | Python only — replaces `MOD_NAME` |
| `HANDLER_FUNCTION_NAME` | `main` | yes | Python only — replaces `FUNC_HANDLER` |
| `HANDLER_PATH` | `./function/handler.js` | yes | Node.js only — unchanged |
| `TRACE_COLLECTOR_ENDPOINT` | `''` | yes | unchanged |
| `PUBLISHER_PROXY_ADDRESS` | `''` | yes | unchanged |

---

## Startup Logging

Printed to stdout before serving requests:

```
Importing function sources from <HANDLER_FOLDER>/<HANDLER_MODULE_NAME>:<HANDLER_FUNCTION_NAME>
Tracing configured with endpoint <TRACE_COLLECTOR_ENDPOINT>
Publisher Proxy available on address <PUBLISHER_PROXY_ADDRESS>
Starting <FUNC_RUNTIME> server <SERVER_HOST>:<SERVER_PORT>
```

---

## Error Handling

| Scenario | Status | Body |
|---|---|---|
| Unhandled exception in user code | 500 | `Internal Server Error` |
| User function not loaded | 500 | `User function not loaded` |
| Request timeout | 408 | (existing behaviour, unchanged) |

Exception details are logged to stderr, never returned to the caller.

---

## File Changes

### `components/runtimes/python314`

| File | Action |
|---|---|
| `server.py` | Rework: new ENVs, startup logging, call `handler.call(method)` with no args passed to user func |
| `lib/module.py` | Rework: call `self.func()` with no args, remove `Event`/CE wiring, 500 on handler error |
| `lib/sdk.py` | **Create**: public SDK |
| `lib/tracing.py` | Fix: remove `__name__ != "__main__"` guard that prevents tracing from initialising when imported |
| `lib/ce.py` | **Delete**: CE logic absorbed into `lib/sdk.py` |

### `components/runtimes/nodejs26`

| File | Action |
|---|---|
| `server.mjs` | Rework: new ENVs, startup logging, pass `(req, res)` to user function |
| `lib/sdk.js` | **Create**: public SDK |
| `lib/helper.js` | Update: `FUNC_TIMEOUT` → `SERVER_CALL_TIMEOUT` |
| `lib/ce.js` | **Delete**: CE logic absorbed into `lib/sdk.js` |

### `components/buildless-serverless/internal/controller/resources/deployment.go`

| Change |
|---|
| `SERVICE_NAMESPACE` → `FUNC_NAMESPACE` in `generalEnvs()` |
| `MOD_NAME` → `HANDLER_MODULE_NAME` in `generalEnvs()` |
| `FUNC_HANDLER` → `HANDLER_FUNCTION_NAME` in `generalEnvs()` |
| `FUNCTION_PATH` → `HANDLER_FOLDER` in `sourceEnvs()` |

No changes needed to `files.go` — the `lib/` directory is already read dynamically, so `lib/sdk.py` and `lib/sdk.js` are automatically included in ejected output.

---

## Testing

Both runtimes have no automated tests. Acceptance is verified manually by running each runtime locally with example handlers covering:

- Basic request/response: `req.body` (Node.js) / `flask.request.get_json()` (Python), return value
- CloudEvent parsing via `getCloudEvent()` / `get_cloud_event()`
- CloudEvent emission via `emitCloudEvent()` / `emit_cloud_event()`
- Tracer access via `getTracer()` / `get_tracer()`
- Context metadata: `getFunctionName()`, `getNamespace()`, `getRuntime()`, `getTimeout()`
- 500 response on handler exception
- Startup log output
- All ENV vars with and without overrides
