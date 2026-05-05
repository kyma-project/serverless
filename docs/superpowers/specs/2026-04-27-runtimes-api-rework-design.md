# Serverless Runtimes API Rework — Design Spec

**Issue:** [#2292](https://github.com/kyma-project/serverless/issues/2292)  
**Date:** 2026-04-27  
**Runtimes in scope:** `python314`, `nodejs26`

---

## Overview

Rework the serverless runtimes API to be more user-friendly, aligned with industry standards, and fully self-contained. The two key changes are:

1. **Handler API (Proposal 3):** User functions take no arguments. Request and response are accessed via an importable SDK module bundled with the runtime.
2. **ENV architecture:** All configuration defaults live in the runtime container. The controller only injects values it cannot calculate itself (function name, namespace, source paths, endpoints).

---

## Components

### python314

| File | Change |
|---|---|
| `server.py` | Reworked: new ENV names, startup logging, no args passed to `main()` |
| `lib/module.py` | Reworked: call user function with no args, 500 on handler error |
| `lib/sdk.py` | **New:** public SDK — `get_request()`, `get_tracer()`, `emit_cloud_event()` |
| `lib/tracing.py` | Minor cleanups only |
| `lib/ce.py` | Removed — CloudEvent logic absorbed into `lib/sdk.py` |

### nodejs26

| File | Change |
|---|---|
| `server.mjs` | Reworked: new ENV names, startup logging, AsyncLocalStorage context store |
| `lib/sdk.js` | **New:** public SDK — `getRequest()`, `getResponse()`, `getTracer()`, `emitCloudEvent()` |
| `lib/helper.js` | Minor cleanups only |
| `lib/ce.js` | Removed — CloudEvent logic absorbed into `lib/sdk.js` |
| `lib/tracer.js` | Kept as-is |
| `lib/metrics.js` | Kept as-is |

### Stub packages (for IDE autocomplete)

Two no-op stub packages that mirror the SDK public API, for use by function developers in their local environments:

- `kyma-sdk` on PyPI — mirrors `lib/sdk.py`
- `@kyma-project/sdk` on npm — mirrors `lib/sdk.js`

These are not required at runtime. Their exact location in the repo and publishing pipeline are out of scope for this issue — only the package content (mirroring the SDK surface) is in scope here.

---

## Handler API

User functions take no arguments. The SDK provides access to the framework request/response objects.

### Python

```python
# handler.py
from sdk import get_request, emit_cloud_event, get_tracer

def main():
    req = get_request()       # flask.Request
    data = req.get_json()
    return {"hello": data}, 200
```

Flask's thread-local context makes `get_request()` safe under concurrent requests without additional wiring — `sdk.get_request()` is a thin wrapper around `flask.request`.

### Node.js

```javascript
// handler.js
import { getRequest, getResponse, emitCloudEvent, getTracer } from 'sdk';

export function main() {
    const req = getRequest();   // Express Request
    const res = getResponse();  // Express Response
    res.json({ hello: req.body });
}
```

The runtime stores `req` and `res` in an `AsyncLocalStorage` store before invoking `main()`. The SDK reads from that store. This is safe under concurrent requests.

### SDK public surface

Same shape in both languages:

| Python | Node.js | Returns | Description |
|---|---|---|---|
| `get_request()` | `getRequest()` | Framework request object | Flask `Request` / Express `Request` |
| _(n/a)_ | `getResponse()` | Express `Response` | Node.js only — Python uses `return` for responses |
| `get_tracer()` | `getTracer()` | OpenTelemetry `Tracer` | Configured tracer instance |
| `emit_cloud_event(type, source, data, attrs?)` | `emitCloudEvent(type, source, data, attrs?)` | — | Publishes a CloudEvent to the publisher proxy |

**Backward compatibility:** Neither runtime had a stable public API. The old `(event, context)` signature is dropped with no compatibility shim.

---

## ENV Architecture

All vars have defaults inside the runtime container. The controller only injects vars it cannot calculate itself.

| ENV var | Default | Injected by |
|---|---|---|
| `FUNC_NAME` | `''` | controller |
| `FUNC_NAMESPACE` | `''` | controller |
| `FUNC_RUNTIME` | `python314` / `nodejs26` | runtime (hardcoded default) |
| `SERVER_HOST` | `0.0.0.0` | runtime |
| `SERVER_PORT` | `8080` | runtime |
| `SERVER_NUMTHREADS` | `50` | runtime (Python only) |
| `SERVER_CALL_TIMEOUT` | `180` | runtime |
| `HANDLER_FOLDER` | `/` | controller |
| `HANDLER_MODULE_NAME` | `handler` | controller (Python only) |
| `HANDLER_FUNCTION_NAME` | `main` | runtime |
| `HANDLER_PATH` | `./handler.js` | controller (Node.js only) |
| `TRACE_COLLECTOR_ENDPOINT` | `''` | controller |
| `PUBLISHER_PROXY_ADDRESS` | `''` | controller |

**Removed vars** (replaced by unified naming):

| Old | Replaced by |
|---|---|
| `SERVICE_NAMESPACE` | `FUNC_NAMESPACE` |
| `FUNC_TIMEOUT` | `SERVER_CALL_TIMEOUT` |
| `REQ_MB_LIMIT` | removed (not exposed, hardcoded default `1`) |
| `KYMA_INTERNAL_LOGGER_ENABLED` | removed |

---

## Startup Logging

Printed to stdout before the server starts accepting requests:

```
Importing function sources from <HANDLER_FOLDER>/<HANDLER_MODULE_NAME>:<HANDLER_FUNCTION_NAME>
Tracing configured with endpoint <TRACE_COLLECTOR_ENDPOINT>
Publisher Proxy available on address <PUBLISHER_PROXY_ADDRESS>
Starting <FUNC_RUNTIME> server <SERVER_HOST>:<SERVER_PORT>
```

---

## Error Handling

| Scenario | HTTP status | Body |
|---|---|---|
| Unhandled exception in user code | 500 | `Internal Server Error` |
| User function not loaded (import failed) | 500 | `User function not loaded` |
| Request timeout | 408 | (existing behaviour, unchanged) |

Exception details are logged to stderr, never leaked to the caller.

---

## Testing

Both runtimes have no automated tests. Acceptance is verified manually by running each runtime locally with example handler functions covering:

- Basic request/response via `get_request()` / `getRequest()`
- CloudEvent emission via `emit_cloud_event()` / `emitCloudEvent()`
- Tracer access via `get_tracer()` / `getTracer()`
- 500 response on handler exception
- Startup log output
- All ENV vars with and without overrides
