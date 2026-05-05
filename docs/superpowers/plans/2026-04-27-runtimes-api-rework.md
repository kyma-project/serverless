# Runtimes API Rework Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rework `python314` and `nodejs26` runtimes so user functions take no arguments and access request/response via an importable SDK module, with unified ENV naming and startup logging.

**Architecture:** Each runtime gets a new `lib/sdk.py` / `lib/sdk.js` that exposes `get_request()` / `getRequest()`, `get_tracer()` / `getTracer()`, and `emit_cloud_event()` / `emitCloudEvent()`. Python uses Flask's thread-local `flask.request`; Node.js stores `req`/`res` in `AsyncLocalStorage` before calling the user function. The old `lib/ce.py` / `lib/ce.js` files are removed — their logic moves into the SDK.

**Tech Stack:** Python 3.14 + Flask + gevent + cloudevents; Node.js 26 + Express 5 + cloudevents; OpenTelemetry for both.

---

## File Map

### python314

| File | Action |
|---|---|
| `components/runtimes/python314/lib/sdk.py` | Create — public SDK |
| `components/runtimes/python314/lib/module.py` | Modify — call handler with no args |
| `components/runtimes/python314/lib/tracing.py` | Modify — fix `__name__` bug |
| `components/runtimes/python314/lib/ce.py` | Delete |
| `components/runtimes/python314/server.py` | Modify — new ENVs, startup log |

### nodejs26

| File | Action |
|---|---|
| `components/runtimes/nodejs26/lib/sdk.js` | Create — public SDK with AsyncLocalStorage |
| `components/runtimes/nodejs26/lib/ce.js` | Delete |
| `components/runtimes/nodejs26/lib/helper.js` | Modify — remove ce import reference |
| `components/runtimes/nodejs26/server.mjs` | Modify — new ENVs, startup log, AsyncLocalStorage wiring |

---

## Task 1: python314 — Create `lib/sdk.py`

**Files:**
- Create: `components/runtimes/python314/lib/sdk.py`

- [ ] **Step 1: Create `lib/sdk.py`**

```python
import requests as _requests
from flask import request as _flask_request
from cloudevents.core.v1.event import CloudEvent
from cloudevents.core.bindings.http import to_structured_event

_tracer = None
_publisher_proxy_address = None


def _configure(tracer, publisher_proxy_address):
    global _tracer, _publisher_proxy_address
    _tracer = tracer
    _publisher_proxy_address = publisher_proxy_address


def get_request():
    return _flask_request


def get_tracer():
    return _tracer


def emit_cloud_event(type, source, data, optional_attributes=None):
    attributes = {"type": type, "source": source}
    if optional_attributes:
        attributes.update(optional_attributes)
    event = CloudEvent(attributes, data)
    message = to_structured_event(event)
    _requests.post(_publisher_proxy_address, data=message.body, headers=message.headers)
```

- [ ] **Step 2: Commit**

```bash
git add components/runtimes/python314/lib/sdk.py
git commit -m "feat(python314): add sdk module"
```

---

## Task 2: python314 — Fix `lib/tracing.py` and rework `lib/module.py`

**Files:**
- Modify: `components/runtimes/python314/lib/tracing.py`
- Modify: `components/runtimes/python314/lib/module.py`

**Context:** `tracing.py` has a bug — `setup()` checks `if __name__ != "__main__"` and returns `None` when imported as a module (which is always). This means tracing never actually initialises. Remove that guard.

`module.py` currently calls `self.func(event, context)`. It needs to call `self.func()` with no arguments instead, and remove all `Event` / CloudEvent wiring (that moves to `sdk.py`).

- [ ] **Step 1: Fix `lib/tracing.py` — remove the `__name__` guard**

Replace the entire `setup()` function body. The file should read:

```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace.export import SimpleSpanProcessor
from opentelemetry.sdk.trace.sampling import DEFAULT_ON
from opentelemetry.instrumentation.requests import RequestsInstrumentor


def setup(tracecollector_endpoint) -> trace.Tracer:
    provider = TracerProvider(
        resource=Resource.create(),
        sampler=DEFAULT_ON,
    )

    if tracecollector_endpoint:
        span_processor = SimpleSpanProcessor(OTLPSpanExporter(endpoint=tracecollector_endpoint))
        provider.add_span_processor(span_processor)

    trace.set_tracer_provider(provider)
    RequestsInstrumentor().instrument()
    return trace.get_tracer("io.kyma-project.serverless")
```

- [ ] **Step 2: Rewrite `lib/module.py`**

Replace the entire file:

```python
import sys
import importlib

import prometheus_client as prom


class Handler:
    def __init__(self, module_folder, module_name, module_function_name):
        sys.path.append(module_folder)
        module = importlib.import_module(module_name)
        self.func = getattr(module, module_function_name)

        self.func_hist = prom.Histogram(
            'function_duration_seconds', 'Duration of user function in seconds', ['method']
        )
        self.func_calls = prom.Counter(
            'function_calls_total', 'Number of calls to user function', ['method']
        )
        self.func_errors = prom.Counter(
            'function_failures_total', 'Number of exceptions in user function', ['method']
        )

    def call(self, method):
        self.func_calls.labels(method).inc()
        with self.func_errors.labels(method).count_exceptions():
            with self.func_hist.labels(method).time():
                return self.func()
```

- [ ] **Step 3: Commit**

```bash
git add components/runtimes/python314/lib/tracing.py components/runtimes/python314/lib/module.py
git commit -m "feat(python314): rework module and fix tracing init"
```

---

## Task 3: python314 — Rework `server.py` and delete `lib/ce.py`

**Files:**
- Modify: `components/runtimes/python314/server.py`
- Delete: `components/runtimes/python314/lib/ce.py`

- [ ] **Step 1: Replace `server.py`**

```python
import os
import sys

import flask
from gevent import pywsgi
import prometheus_client

from lib import tracing, module, sdk

func_name = os.getenv('FUNC_NAME', '')
func_namespace = os.getenv('FUNC_NAMESPACE', '')
func_runtime = os.getenv('FUNC_RUNTIME', 'python314')
server_host = os.getenv('SERVER_HOST', '0.0.0.0')
server_port = int(os.getenv('SERVER_PORT', '8080'))
server_numthreads = int(os.getenv('SERVER_NUMTHREADS', '50'))
server_call_timeout = float(os.getenv('SERVER_CALL_TIMEOUT', '180'))
handler_folder = os.getenv('HANDLER_FOLDER', '/')
handler_module_name = os.getenv('HANDLER_MODULE_NAME', 'handler')
handler_function_name = os.getenv('HANDLER_FUNCTION_NAME', 'main')
trace_collector_endpoint = os.getenv('TRACE_COLLECTOR_ENDPOINT', '')
publisher_proxy_address = os.getenv('PUBLISHER_PROXY_ADDRESS', '')

print(f"Importing function sources from {handler_folder}/{handler_module_name}:{handler_function_name}")
print(f"Tracing configured with endpoint {trace_collector_endpoint}")
print(f"Publisher Proxy available on address {publisher_proxy_address}")
print(f"Starting {func_runtime} server {server_host}:{server_port}")
sys.stdout.flush()

tracer = tracing.setup(trace_collector_endpoint)
sdk._configure(tracer, publisher_proxy_address)

handler = module.Handler(handler_folder, handler_module_name, handler_function_name)

app = flask.Flask(__name__)


@app.route('/', methods=['GET', 'POST', 'PUT', 'HEAD', 'OPTIONS', 'DELETE'])
def userfunc_call():
    return handler.call(flask.request.method)


@app.errorhandler(500)
def internal_error(error):
    return 'Internal Server Error', 500


@app.get('/favicon.ico')
def favicon():
    return '', 204


@app.get('/healthz')
def healthz():
    return 'OK', 200


@app.get('/metrics')
def metrics():
    return prometheus_client.generate_latest(prometheus_client.REGISTRY), 200, {'Content-Type': prometheus_client.CONTENT_TYPE_LATEST}


if __name__ == "__main__":
    pywsgi.WSGIServer(
        (server_host, server_port),
        app,
        spawn=server_numthreads,
        log=None,
    ).serve_forever()
```

- [ ] **Step 2: Delete `lib/ce.py`**

```bash
rm components/runtimes/python314/lib/ce.py
```

- [ ] **Step 3: Commit**

```bash
git add components/runtimes/python314/server.py
git rm components/runtimes/python314/lib/ce.py
git commit -m "feat(python314): rework server, unified ENVs, startup logging"
```

---

## Task 4: python314 — Manual verification

**No automated tests exist for this runtime. Verify manually.**

- [ ] **Step 1: Create a temporary test handler**

Create `/tmp/handler.py`:

```python
from sdk import get_request, get_tracer, emit_cloud_event

def main():
    req = get_request()
    return {"method": req.method, "body": req.get_json(silent=True)}, 200
```

- [ ] **Step 2: Run the server locally**

```bash
cd components/runtimes/python314
HANDLER_FOLDER=/tmp HANDLER_MODULE_NAME=handler HANDLER_FUNCTION_NAME=main python server.py
```

Expected stdout before serving:
```
Importing function sources from /tmp/handler:main
Tracing configured with endpoint 
Publisher Proxy available on address 
Starting python314 server 0.0.0.0:8080
```

- [ ] **Step 3: Smoke test basic request**

```bash
curl -s -X POST http://localhost:8080 -H 'Content-Type: application/json' -d '{"hello":"world"}'
```

Expected: `{"body":{"hello":"world"},"method":"POST"}` with HTTP 200.

- [ ] **Step 4: Smoke test 500 on handler error**

Create `/tmp/broken_handler.py`:

```python
def main():
    raise RuntimeError("boom")
```

Restart the server with `HANDLER_MODULE_NAME=broken_handler`, then:

```bash
curl -sv http://localhost:8080
```

Expected: HTTP 500, body `Internal Server Error`.

- [ ] **Step 5: Commit verification note (no code change needed)**

```bash
git commit --allow-empty -m "chore(python314): manual verification passed"
```

---

## Task 5: nodejs26 — Create `lib/sdk.js`

**Files:**
- Create: `components/runtimes/nodejs26/lib/sdk.js`

The SDK uses Node.js `AsyncLocalStorage` (built-in, no extra dependency). The server calls `runWithContext(req, res, fn)` before invoking user code; the SDK's `getRequest()` / `getResponse()` read from that store.

- [ ] **Step 1: Create `lib/sdk.js`**

```javascript
'use strict';

const { AsyncLocalStorage } = require('async_hooks');
const { HTTP, CloudEvent } = require('cloudevents');
const axios = require('axios');

const _store = new AsyncLocalStorage();

let _tracer = null;
let _publisherProxyAddress = null;

function _configure(tracer, publisherProxyAddress) {
    _tracer = tracer;
    _publisherProxyAddress = publisherProxyAddress;
}

function runWithContext(req, res, fn) {
    return _store.run({ req, res }, fn);
}

function getRequest() {
    const ctx = _store.getStore();
    if (!ctx) throw new Error('getRequest() called outside of a request context');
    return ctx.req;
}

function getResponse() {
    const ctx = _store.getStore();
    if (!ctx) throw new Error('getResponse() called outside of a request context');
    return ctx.res;
}

function getTracer() {
    return _tracer;
}

function emitCloudEvent(type, source, data, optionalAttributes) {
    const attrs = Object.assign({ type, source, data }, optionalAttributes || {});
    if (!attrs.datacontenttype) {
        attrs.datacontenttype = typeof data === 'object' ? 'application/json' : 'text/plain';
    }
    const ce = new CloudEvent(attrs);
    const message = HTTP.structured(ce);
    return axios.post(_publisherProxyAddress, message.body, { headers: message.headers });
}

module.exports = { _configure, runWithContext, getRequest, getResponse, getTracer, emitCloudEvent };
```

- [ ] **Step 2: Commit**

```bash
git add components/runtimes/nodejs26/lib/sdk.js
git commit -m "feat(nodejs26): add sdk module with AsyncLocalStorage"
```

---

## Task 6: nodejs26 — Rework `server.mjs` and delete `lib/ce.js`

**Files:**
- Modify: `components/runtimes/nodejs26/server.mjs`
- Delete: `components/runtimes/nodejs26/lib/ce.js`

Note: `lib/tracer.js` and `lib/metrics.js` are kept unchanged. `lib/helper.js` no longer needs the `ce` import (it never imported it — `ce` was only used in `server.mjs`).

- [ ] **Step 1: Replace `server.mjs`**

```javascript
import sdk from './lib/sdk.js';
import helper from './lib/helper.js';
import bodyParser from 'body-parser';
import process from 'process';

import { setupTracer, getCurrentSpan } from './lib/tracer.js';
import { getMetrics, setupMetrics, createFunctionDurationHistogram, createFunctionCallsTotalCounter, createFunctionFailuresTotalCounter } from './lib/metrics.js';

process.on("uncaughtException", (err) => {
    console.error(`Caught exception: ${err}`);
});

const funcName = process.env.FUNC_NAME || '';
const funcNamespace = process.env.FUNC_NAMESPACE || '';
const funcRuntime = process.env.FUNC_RUNTIME || 'nodejs26';
const serverHost = process.env.SERVER_HOST || '0.0.0.0';
const serverPort = parseInt(process.env.SERVER_PORT || '8080', 10);
const serverCallTimeout = Number(process.env.SERVER_CALL_TIMEOUT || '180');
const handlerPath = process.env.HANDLER_PATH || './handler.js';
const traceCollectorEndpoint = process.env.TRACE_COLLECTOR_ENDPOINT || '';
const publisherProxyAddress = process.env.PUBLISHER_PROXY_ADDRESS || '';

console.log(`Importing function sources from ${handlerPath}:main`);
console.log(`Tracing configured with endpoint ${traceCollectorEndpoint}`);
console.log(`Publisher Proxy available on address ${publisherProxyAddress}`);
console.log(`Starting ${funcRuntime} server ${serverHost}:${serverPort}`);

const tracer = setupTracer(funcName);
setupMetrics(funcName);
sdk._configure(tracer, publisherProxyAddress);

const callsTotalCounter = createFunctionCallsTotalCounter(funcName);
const failuresTotalCounter = createFunctionFailuresTotalCounter(funcName);
const durationHistogram = createFunctionDurationHistogram(funcName);

// require express AFTER tracer setup
import express from "express";
const app = express();

let userFunction;

app.use(bodyParser.json({ type: ['application/json', 'application/cloudevents+json'], limit: '1mb', strict: false }));
app.use(bodyParser.text({ type: ['text/*'], limit: '1mb' }));
app.use(bodyParser.urlencoded({ limit: '1mb', extended: true }));
app.use(bodyParser.raw({ limit: '1mb', type: () => true }));

app.use(helper.handleTimeOut);

app.get("/healthz", (req, res) => res.status(200).send("OK"));
app.get("/metrics", (req, res) => getMetrics(req, res));
app.get('/favicon.ico', (req, res) => res.status(204).end());

app.all("*path", (req, res) => {
    res.header('Access-Control-Allow-Origin', '*');

    if (req.method === 'OPTIONS') {
        res.header('Access-Control-Allow-Methods', req.headers['access-control-request-method']);
        res.header('Access-Control-Allow-Headers', req.headers['access-control-request-headers']);
        res.end();
        return;
    }

    callsTotalCounter.add(1);
    const startTime = new Date().getTime();

    if (!userFunction) {
        failuresTotalCounter.add(1);
        res.status(500).send("User function not loaded");
        return;
    }

    const currentSpan = getCurrentSpan();

    sdk.runWithContext(req, res, () => {
        try {
            const out = userFunction();
            if (out && helper.isPromise(out)) {
                out.then(result => {
                    if (!res.writableEnded) res.json(result);
                }).catch(err => {
                    helper.handleError(err, currentSpan, (body, status) => {
                        if (!res.writableEnded) res.status(status || 500).send(body);
                    });
                    failuresTotalCounter.add(1);
                });
            } else if (out !== undefined && !res.writableEnded) {
                res.json(out);
            }
        } catch (err) {
            helper.handleError(err, currentSpan, (body, status) => {
                if (!res.writableEnded) res.status(status || 500).send(body);
            });
            failuresTotalCounter.add(1);
        }
    });

    const endTime = new Date().getTime();
    durationHistogram.record(endTime - startTime);
});

const server = app.listen(serverPort, serverHost);
helper.configureGracefulShutdown(server);

let startTime = process.hrtime();
import(handlerPath).then((fn) => {
    if (helper.isFunction(fn.main)) {
        userFunction = fn.main;
        const elapsed = process.hrtime(startTime);
        console.log(`user code loaded in ${elapsed[0]}sec ${elapsed[1] / 1000000}ms`);
    } else {
        console.error("Content loaded is not a function. Make sure your function exports 'main' function", fn);
    }
}).catch((err) => {
    console.error("Failed to load user function:", err);
});
```

- [ ] **Step 2: Delete `lib/ce.js`**

```bash
rm components/runtimes/nodejs26/lib/ce.js
```

- [ ] **Step 3: Update `lib/helper.js` — replace `handleTimeOut` to use `SERVER_CALL_TIMEOUT`**

In `lib/helper.js`, the `handleTimeOut` function reads `FUNC_TIMEOUT`. Update it to read `SERVER_CALL_TIMEOUT` to match the new ENV naming:

```javascript
function handleTimeOut(req, res, next){
  const timeout = Number(process.env.SERVER_CALL_TIMEOUT || '180');
  res.setTimeout(timeout * 1000, function(){
    res.sendStatus(408);
  });
  next();
}
```

- [ ] **Step 4: Commit**

```bash
git add components/runtimes/nodejs26/server.mjs components/runtimes/nodejs26/lib/helper.js
git rm components/runtimes/nodejs26/lib/ce.js
git commit -m "feat(nodejs26): rework server, SDK wiring, unified ENVs, startup logging"
```

---

## Task 7: nodejs26 — Manual verification

**No automated tests exist for this runtime. Verify manually.**

- [ ] **Step 1: Install dependencies**

```bash
cd components/runtimes/nodejs26
npm install
```

- [ ] **Step 2: Create a temporary test handler**

Create `/tmp/handler.mjs`. The SDK is resolved relative to `NODE_PATH`, which the server sets via the runtime working directory — but for local testing the simplest approach is to use `getRequest` from the absolute path:

```javascript
import { createRequire } from 'module';
import { fileURLToPath } from 'url';
import path from 'path';

const require = createRequire(import.meta.url);
// Adjust path to wherever your local checkout lives
const sdk = require('/Users/i542853/go/src/github.com/kyma-project/serverless/components/runtimes/nodejs26/lib/sdk.js');

export function main() {
    const req = sdk.getRequest();
    return { method: req.method, body: req.body };
}
```

- [ ] **Step 3: Run the server locally**

```bash
cd components/runtimes/nodejs26
HANDLER_PATH=/tmp/handler.mjs node server.mjs
```

Expected stdout:
```
Importing function sources from /tmp/handler.mjs:main
Tracing configured with endpoint 
Publisher Proxy available on address 
Starting nodejs26 server 0.0.0.0:8080
```

- [ ] **Step 4: Smoke test basic request**

```bash
curl -s -X POST http://localhost:8080 -H 'Content-Type: application/json' -d '{"hello":"world"}'
```

Expected: `{"method":"POST","body":{"hello":"world"}}` with HTTP 200.

- [ ] **Step 5: Smoke test 500 on handler error**

Create `/tmp/broken_handler.mjs`:

```javascript
export function main() {
    throw new Error("boom");
}
```

Restart the server with `HANDLER_PATH=/tmp/broken_handler.mjs`, then:

```bash
curl -sv http://localhost:8080
```

Expected: HTTP 500, body `Internal Server Error` (from `helper.handleError` → `resolveErrorMsg`).

- [ ] **Step 6: Commit verification note**

```bash
git commit --allow-empty -m "chore(nodejs26): manual verification passed"
```

---

## Task 8: Stub packages

**Files:**
- Create: `components/runtimes/python314/stub/sdk.py`
- Create: `components/runtimes/nodejs26/stub/sdk.js`

These are no-op stubs that mirror the real SDK surface. Function developers install them locally for IDE autocomplete. They are never used at runtime.

- [ ] **Step 1: Create Python stub at `components/runtimes/python314/stub/sdk.py`**

```python
"""
Stub SDK for local development and IDE autocomplete.
At runtime, the real sdk module bundled with the python314 container is used instead.

Install: pip install kyma-sdk
"""
from flask import Request


def get_request() -> Request:
    """Returns the current Flask request object."""
    raise RuntimeError("sdk stub: not available outside a running Kyma function")


def get_tracer():
    """Returns the configured OpenTelemetry tracer."""
    raise RuntimeError("sdk stub: not available outside a running Kyma function")


def emit_cloud_event(type: str, source: str, data, optional_attributes: dict = None):
    """Publishes a CloudEvent to the configured publisher proxy."""
    raise RuntimeError("sdk stub: not available outside a running Kyma function")
```

- [ ] **Step 2: Create Node.js stub at `components/runtimes/nodejs26/stub/sdk.js`**

```javascript
/**
 * Stub SDK for local development and IDE autocomplete.
 * At runtime, the real sdk module bundled with the nodejs26 container is used instead.
 *
 * Install: npm install @kyma-project/sdk
 */

/** @returns {import('express').Request} */
function getRequest() {
    throw new Error('sdk stub: not available outside a running Kyma function');
}

/** @returns {import('express').Response} */
function getResponse() {
    throw new Error('sdk stub: not available outside a running Kyma function');
}

/** @returns {import('@opentelemetry/api').Tracer} */
function getTracer() {
    throw new Error('sdk stub: not available outside a running Kyma function');
}

/**
 * @param {string} type
 * @param {string} source
 * @param {*} data
 * @param {object} [optionalAttributes]
 */
function emitCloudEvent(type, source, data, optionalAttributes) {
    throw new Error('sdk stub: not available outside a running Kyma function');
}

module.exports = { getRequest, getResponse, getTracer, emitCloudEvent };
```

- [ ] **Step 3: Commit**

```bash
git add components/runtimes/python314/stub/sdk.py components/runtimes/nodejs26/stub/sdk.js
git commit -m "feat: add SDK stub packages for IDE autocomplete"
```
