# Runtimes API Rework — Proposal 4 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rework `python314` and `nodejs26` runtimes so Python handlers take no args (Flask-native) and Node.js handlers receive `(req, res)` directly (Express-native), with a shared SDK module replacing the old `event`/`context` arguments.

**Architecture:** Each runtime gets a new `lib/sdk.py` / `lib/sdk.js` configured once at startup via `_configure(...)`. Python's `server.py` calls `handler.call(method)` which calls `self.func()` with no args. Node.js `server.mjs` calls `userFunction(req, res)` directly. The old `lib/ce.js` is deleted; its CloudEvent logic moves into `lib/sdk.js`. The controller's `deployment.go` is updated to rename three ENV vars.

**Tech Stack:** Python 3.14 + Flask + gevent + cloudevents; Node.js 26 + Express 5 + cloudevents; OpenTelemetry for both.

---

## File Map

### python314

| File | Action |
|---|---|
| `components/runtimes/python314/lib/sdk.py` | **Create** — public SDK: tracer, CloudEvent helpers, function metadata |
| `components/runtimes/python314/lib/tracing.py` | **Modify** — remove `__name__ != "__main__"` guard that silently returns `None` |
| `components/runtimes/python314/lib/module.py` | **Modify** — call `self.func()` with no args, remove `Event`/CE wiring |
| `components/runtimes/python314/server.py` | **Modify** — new ENVs, startup logging, wire `sdk._configure(...)` |

### nodejs26

| File | Action |
|---|---|
| `components/runtimes/nodejs26/lib/sdk.js` | **Create** — public SDK: tracer, CloudEvent helpers, function metadata |
| `components/runtimes/nodejs26/lib/ce.js` | **Delete** — CE logic absorbed into `lib/sdk.js` |
| `components/runtimes/nodejs26/lib/helper.js` | **Modify** — `FUNC_TIMEOUT` → `SERVER_CALL_TIMEOUT` |
| `components/runtimes/nodejs26/server.mjs` | **Modify** — new ENVs, startup logging, pass `(req, res)` to user function |

### controller

| File | Action |
|---|---|
| `components/buildless-serverless/internal/controller/resources/deployment.go` | **Modify** — rename three ENV vars in `generalEnvs()` and `sourceEnvs()` |

---

## Task 1: python314 — Fix `lib/tracing.py`

**Files:**
- Modify: `components/runtimes/python314/lib/tracing.py`

The current `setup()` function has `if __name__ != "__main__": return None` at the top. Since this file is always imported (never run directly), `__name__` is always `"lib.tracing"` — so tracing silently never initialises. Remove that guard. Also remove the now-unused `set_req_context` helper (it was only needed for the old thread-based request dispatch, which Flask's built-in context replaces).

- [ ] **Step 1: Replace `lib/tracing.py`**

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

- [ ] **Step 2: Commit**

```bash
git add components/runtimes/python314/lib/tracing.py
git commit -m "fix(python314): remove __name__ guard that prevented tracing from initialising"
```

---

## Task 2: python314 — Create `lib/sdk.py`

**Files:**
- Create: `components/runtimes/python314/lib/sdk.py`

The SDK stores server-level config as module-level variables set once at startup by `_configure()`. `get_cloud_event()` reads Flask's thread-local `flask.request` directly — safe because Flask guarantees per-request thread isolation.

- [ ] **Step 1: Create `lib/sdk.py`**

```python
import requests as _requests
from flask import request as _flask_request
from cloudevents.core.v1.event import CloudEvent
from cloudevents.core.bindings.http import from_http_event, to_structured_event, HTTPMessage

_tracer = None
_publisher_proxy_address = None
_func_name = ''
_func_namespace = ''
_func_runtime = ''
_server_call_timeout = 180.0


def _configure(tracer, publisher_proxy_address, func_name, func_namespace, func_runtime, server_call_timeout):
    global _tracer, _publisher_proxy_address, _func_name, _func_namespace, _func_runtime, _server_call_timeout
    _tracer = tracer
    _publisher_proxy_address = publisher_proxy_address
    _func_name = func_name
    _func_namespace = func_namespace
    _func_runtime = func_runtime
    _server_call_timeout = server_call_timeout


def get_tracer():
    return _tracer


def get_cloud_event():
    req = _flask_request
    content_type = req.content_type or ''
    has_ce_content_type = 'application/cloudevents+json' in content_type.split(';')
    has_ce_headers = 'ce-type' in req.headers and 'ce-source' in req.headers
    if not (has_ce_content_type or has_ce_headers):
        return None
    message = HTTPMessage(headers=dict(req.headers), body=req.get_data())
    return from_http_event(message)


def emit_cloud_event(type, source, data, optional_attributes=None):
    attributes = {'type': type, 'source': source}
    if optional_attributes:
        attributes.update(optional_attributes)
    event = CloudEvent(attributes, data)
    message = to_structured_event(event)
    _requests.post(_publisher_proxy_address, data=message.body, headers=message.headers)


def get_function_name():
    return _func_name


def get_namespace():
    return _func_namespace


def get_runtime():
    return _func_runtime


def get_timeout():
    return _server_call_timeout
```

- [ ] **Step 2: Commit**

```bash
git add components/runtimes/python314/lib/sdk.py
git commit -m "feat(python314): add sdk module"
```

---

## Task 3: python314 — Rework `lib/module.py`

**Files:**
- Modify: `components/runtimes/python314/lib/module.py`

Remove the `Event` class and all CE/request wiring. `Handler.call(method)` now calls `self.func()` with no arguments — Flask's thread-local context makes the request available via `flask.request` inside the user function.

- [ ] **Step 1: Replace `lib/module.py`**

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

- [ ] **Step 2: Commit**

```bash
git add components/runtimes/python314/lib/module.py
git commit -m "feat(python314): rework module — call handler with no args"
```

---

## Task 4: python314 — Rework `server.py`

**Files:**
- Modify: `components/runtimes/python314/server.py`

New ENVs with unified naming, startup logging, wires `sdk._configure(...)` before serving. The Flask app calls `handler.call(flask.request.method)` — the user function runs in the Flask request context so `flask.request` is available inside it.

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
handler_folder = os.getenv('HANDLER_FOLDER', '/kubeless')
handler_module_name = os.getenv('HANDLER_MODULE_NAME', 'handler')
handler_function_name = os.getenv('HANDLER_FUNCTION_NAME', 'main')
trace_collector_endpoint = os.getenv('TRACE_COLLECTOR_ENDPOINT', '')
publisher_proxy_address = os.getenv('PUBLISHER_PROXY_ADDRESS', '')

print(f"Importing function sources from {handler_folder}/{handler_module_name}:{handler_function_name}", flush=True)
print(f"Tracing configured with endpoint {trace_collector_endpoint}", flush=True)
print(f"Publisher Proxy available on address {publisher_proxy_address}", flush=True)
print(f"Starting {func_runtime} server {server_host}:{server_port}", flush=True)

tracer = tracing.setup(trace_collector_endpoint)
sdk._configure(tracer, publisher_proxy_address, func_name, func_namespace, func_runtime, server_call_timeout)

handler = module.Handler(handler_folder, handler_module_name, handler_function_name)

app = flask.Flask(__name__)


@app.route('/', defaults={'path': ''}, methods=['GET', 'POST', 'PUT', 'HEAD', 'OPTIONS', 'DELETE', 'PATCH'])
@app.route('/<path:path>', methods=['GET', 'POST', 'PUT', 'HEAD', 'OPTIONS', 'DELETE', 'PATCH'])
def userfunc_call(path=''):
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


if __name__ == '__main__':
    pywsgi.WSGIServer(
        (server_host, server_port),
        app,
        spawn=server_numthreads,
        log=None,
    ).serve_forever()
```

- [ ] **Step 2: Commit**

```bash
git add components/runtimes/python314/server.py
git commit -m "feat(python314): rework server — new ENVs, startup logging, no-args handler"
```

---

## Task 5: python314 — Manual verification

No automated tests exist. Verify manually by running the server locally.

- [ ] **Step 1: Install dependencies**

```bash
cd components/runtimes/python314
python3 -m venv --without-scm-ignore-files .venv
.venv/bin/pip install -r requirements.txt
```

- [ ] **Step 2: Create a test handler at `/tmp/handler.py`**

```python
from flask import request
import sdk

def main():
    return {
        "method": request.method,
        "body": request.get_json(silent=True),
        "runtime": sdk.get_runtime(),
        "name": sdk.get_function_name(),
        "namespace": sdk.get_namespace(),
        "timeout": sdk.get_timeout(),
    }, 200
```

- [ ] **Step 3: Run the server**

```bash
cd components/runtimes/python314
HANDLER_FOLDER=/tmp HANDLER_MODULE_NAME=handler HANDLER_FUNCTION_NAME=main FUNC_NAME=test-fn FUNC_NAMESPACE=test-ns .venv/bin/python server.py
```

Expected stdout before serving:
```
Importing function sources from /tmp/handler:main
Tracing configured with endpoint 
Publisher Proxy available on address 
Starting python314 server 0.0.0.0:8080
```

- [ ] **Step 4: Smoke test — basic request**

```bash
curl -s -X POST http://localhost:8080 -H 'Content-Type: application/json' -d '{"hello":"world"}'
```

Expected: HTTP 200, body:
```json
{"body":{"hello":"world"},"method":"POST","name":"test-fn","namespace":"test-ns","runtime":"python314","timeout":180.0}
```

- [ ] **Step 5: Smoke test — 500 on handler exception**

Create `/tmp/broken_handler.py`:
```python
def main():
    raise RuntimeError("boom")
```

Restart server with `HANDLER_MODULE_NAME=broken_handler`, then:
```bash
curl -sv http://localhost:8080
```

Expected: HTTP 500, body `Internal Server Error`.

- [ ] **Step 6: Smoke test — startup logging with custom ENVs**

```bash
SERVER_PORT=9090 FUNC_RUNTIME=python314-custom TRACE_COLLECTOR_ENDPOINT=http://jaeger:4318 \
  HANDLER_FOLDER=/tmp HANDLER_MODULE_NAME=handler HANDLER_FUNCTION_NAME=main \
  .venv/bin/python server.py 2>&1 | head -4
```

Expected:
```
Importing function sources from /tmp/handler:main
Tracing configured with endpoint http://jaeger:4318
Publisher Proxy available on address 
Starting python314-custom server 0.0.0.0:9090
```

- [ ] **Step 7: Commit verification note**

```bash
git commit --allow-empty -m "chore(python314): manual verification passed"
```

---

## Task 6: nodejs26 — Create `lib/sdk.js` and delete `lib/ce.js`

**Files:**
- Create: `components/runtimes/nodejs26/lib/sdk.js`
- Delete: `components/runtimes/nodejs26/lib/ce.js`

The SDK stores server-level config as module-level variables. `getCloudEvent(req)` receives the current `req` as a parameter — unlike Python's Flask there is no framework-level thread-local in Express, so the caller passes `req` explicitly. This means the Node.js signature is `sdk.getCloudEvent(req)` while the Python signature is `sdk.get_cloud_event()` (reads `flask.request` internally).

- [ ] **Step 1: Create `lib/sdk.js`**

```javascript
'use strict';

const { HTTP, CloudEvent } = require('cloudevents');
const axios = require('axios');

let _tracer = null;
let _publisherProxyAddress = null;
let _funcName = '';
let _funcNamespace = '';
let _funcRuntime = '';
let _serverCallTimeout = 180;

function _configure(tracer, publisherProxyAddress, funcName, funcNamespace, funcRuntime, serverCallTimeout) {
    _tracer = tracer;
    _publisherProxyAddress = publisherProxyAddress;
    _funcName = funcName;
    _funcNamespace = funcNamespace;
    _funcRuntime = funcRuntime;
    _serverCallTimeout = serverCallTimeout;
}

function getTracer() {
    return _tracer;
}

function getCloudEvent(req) {
    const isCloudEventContentType = (req.get('content-type') || '').includes('application/cloudevents+json');
    const hasCeHeaders = req.get('ce-type') && req.get('ce-source');
    if (!isCloudEventContentType && !hasCeHeaders) {
        return null;
    }
    try {
        return HTTP.toEvent({ headers: req.headers, body: req.body });
    } catch (e) {
        return null;
    }
}

function emitCloudEvent(type, source, data, optionalAttributes) {
    const attrs = Object.assign({ type, source }, optionalAttributes || {});
    if (!attrs.datacontenttype) {
        attrs.datacontenttype = typeof data === 'object' ? 'application/json' : 'text/plain';
    }
    const ce = new CloudEvent(Object.assign(attrs, { data }));
    const message = HTTP.structured(ce);
    return axios.post(_publisherProxyAddress, message.body, { headers: message.headers });
}

function getFunctionName() { return _funcName; }
function getNamespace()    { return _funcNamespace; }
function getRuntime()      { return _funcRuntime; }
function getTimeout()      { return _serverCallTimeout; }

module.exports = { _configure, getTracer, getCloudEvent, emitCloudEvent, getFunctionName, getNamespace, getRuntime, getTimeout };
```

- [ ] **Step 2: Delete `lib/ce.js`**

```bash
git rm components/runtimes/nodejs26/lib/ce.js
```

- [ ] **Step 3: Commit**

```bash
git add components/runtimes/nodejs26/lib/sdk.js
git commit -m "feat(nodejs26): add sdk module, delete ce.js"
```

---

## Task 7: nodejs26 — Update `lib/helper.js`

**Files:**
- Modify: `components/runtimes/nodejs26/lib/helper.js`

One change: `handleTimeOut` reads `FUNC_TIMEOUT` — rename to `SERVER_CALL_TIMEOUT`.

- [ ] **Step 1: Edit `handleTimeOut` in `lib/helper.js`**

Find this function (around line 18):
```javascript
function handleTimeOut(req, res, next){
  const timeout = Number(process.env.FUNC_TIMEOUT || '180');
  res.setTimeout(timeout*1000, function(){
          res.sendStatus(408);
      });
  next();
}
```

Replace with:
```javascript
function handleTimeOut(req, res, next){
  const timeout = Number(process.env.SERVER_CALL_TIMEOUT || '180');
  res.setTimeout(timeout * 1000, function(){
    res.sendStatus(408);
  });
  next();
}
```

- [ ] **Step 2: Commit**

```bash
git add components/runtimes/nodejs26/lib/helper.js
git commit -m "fix(nodejs26): rename FUNC_TIMEOUT to SERVER_CALL_TIMEOUT in handleTimeOut"
```

---

## Task 8: nodejs26 — Rework `server.mjs`

**Files:**
- Modify: `components/runtimes/nodejs26/server.mjs`

Key changes from current `server.mjs`:
- Remove `ce` import (deleted)
- New ENV names with defaults
- Startup logging
- Wire `sdk._configure(...)`
- Call `userFunction(req, res)` instead of `userFunction(event, context, sendResponse)`
- Remove `sendResponse`, `buildEvent`, `context` object construction
- Pass `req` to `sdk.getCloudEvent(req)` — Node.js has no framework thread-local

- [ ] **Step 1: Replace `server.mjs`**

```javascript
import sdk from './lib/sdk.js';
import helper from './lib/helper.js';
import bodyParser from 'body-parser';
import process from 'process';

import { setupTracer, getCurrentSpan } from './lib/tracer.js';
import { getMetrics, setupMetrics, createFunctionDurationHistogram, createFunctionCallsTotalCounter, createFunctionFailuresTotalCounter } from './lib/metrics.js';

process.on('uncaughtException', (err) => {
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
sdk._configure(tracer, publisherProxyAddress, funcName, funcNamespace, funcRuntime, serverCallTimeout);

const callsTotalCounter = createFunctionCallsTotalCounter(funcName);
const failuresTotalCounter = createFunctionFailuresTotalCounter(funcName);
const durationHistogram = createFunctionDurationHistogram(funcName);

// express must be imported AFTER tracer setup
import express from 'express';
const app = express();

let userFunction;

app.use(bodyParser.json({ type: ['application/json', 'application/cloudevents+json'], limit: '1mb', strict: false }));
app.use(bodyParser.text({ type: ['text/*'], limit: '1mb' }));
app.use(bodyParser.urlencoded({ limit: '1mb', extended: true }));
app.use(bodyParser.raw({ limit: '1mb', type: () => true }));

app.use(helper.handleTimeOut);

app.get('/healthz', (req, res) => res.status(200).send('OK'));
app.get('/metrics', (req, res) => getMetrics(req, res));
app.get('/favicon.ico', (req, res) => res.status(204).end());

app.all('*path', (req, res) => {
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
        res.status(500).send('User function not loaded');
        return;
    }

    const currentSpan = getCurrentSpan();

    try {
        const out = userFunction(req, res);
        if (out && helper.isPromise(out)) {
            out.catch((err) => {
                failuresTotalCounter.add(1);
                helper.handleError(err, currentSpan, (body, status) => {
                    if (!res.writableEnded) res.status(status || 500).send(body);
                });
            });
        }
    } catch (err) {
        failuresTotalCounter.add(1);
        helper.handleError(err, currentSpan, (body, status) => {
            if (!res.writableEnded) res.status(status || 500).send(body);
        });
    }

    const endTime = new Date().getTime();
    durationHistogram.record(endTime - startTime);
});

const server = app.listen(serverPort, serverHost);
helper.configureGracefulShutdown(server);

const startTime = process.hrtime();
import(handlerPath).then((fn) => {
    if (helper.isFunction(fn.main)) {
        userFunction = fn.main;
        const elapsed = process.hrtime(startTime);
        console.log(`user code loaded in ${elapsed[0]}sec ${elapsed[1] / 1000000}ms`);
    } else {
        console.error("Content loaded is not a function. Make sure your function exports 'main' function", fn);
    }
}).catch((err) => {
    console.error('Failed to load user function:', err);
});
```

- [ ] **Step 2: Commit**

```bash
git add components/runtimes/nodejs26/server.mjs
git commit -m "feat(nodejs26): rework server — new ENVs, startup logging, (req, res) handler"
```

---

## Task 9: nodejs26 — Manual verification

No automated tests exist. Verify manually.

- [ ] **Step 1: Install dependencies**

```bash
cd components/runtimes/nodejs26
npm install
```

- [ ] **Step 2: Create a test handler at `/tmp/handler.mjs`**

```javascript
import sdk from '/path/to/components/runtimes/nodejs26/lib/sdk.js';

export function main(req, res) {
    res.json({
        method: req.method,
        body: req.body,
        runtime: sdk.getRuntime(),
        name: sdk.getFunctionName(),
        namespace: sdk.getNamespace(),
        timeout: sdk.getTimeout(),
        cloudEvent: sdk.getCloudEvent(req),
    });
}
```

Replace `/path/to/` with the absolute path to your checkout, e.g. `/Users/i542853/go/src/github.com/kyma-project/serverless`.

- [ ] **Step 3: Run the server**

```bash
cd components/runtimes/nodejs26
HANDLER_PATH=/tmp/handler.mjs FUNC_NAME=test-fn FUNC_NAMESPACE=test-ns node server.mjs
```

Expected stdout:
```
Importing function sources from /tmp/handler.mjs:main
Tracing configured with endpoint 
Publisher Proxy available on address 
Starting nodejs26 server 0.0.0.0:8080
user code loaded in 0sec ...ms
```

- [ ] **Step 4: Smoke test — basic request**

```bash
curl -s -X POST http://localhost:8080 -H 'Content-Type: application/json' -d '{"hello":"world"}'
```

Expected: HTTP 200, body:
```json
{"method":"POST","body":{"hello":"world"},"runtime":"nodejs26","name":"test-fn","namespace":"test-ns","timeout":180,"cloudEvent":null}
```

- [ ] **Step 5: Smoke test — incoming CloudEvent**

```bash
curl -s -X POST http://localhost:8080 \
  -H 'Content-Type: application/cloudevents+json' \
  -d '{"specversion":"1.0","type":"com.example.test","source":"/test","id":"abc","data":{"msg":"hi"}}'
```

Expected: HTTP 200, `cloudEvent` field is non-null with `type`, `source`, `id` populated.

- [ ] **Step 6: Smoke test — 500 on handler exception**

Create `/tmp/broken_handler.mjs`:
```javascript
export function main(req, res) {
    throw new Error('boom');
}
```

Restart server with `HANDLER_PATH=/tmp/broken_handler.mjs`, then:
```bash
curl -sv http://localhost:8080
```

Expected: HTTP 500, body `Internal server error`.

- [ ] **Step 7: Commit verification note**

```bash
git commit --allow-empty -m "chore(nodejs26): manual verification passed"
```

---

## Task 10: Controller — Rename ENV vars in `deployment.go`

**Files:**
- Modify: `components/buildless-serverless/internal/controller/resources/deployment.go`

Three renames in `generalEnvs()` and one in `sourceEnvs()`. No logic changes.

- [ ] **Step 1: In `generalEnvs()` (around line 653), rename three ENV vars**

Find:
```go
		{
			Name:  "SERVICE_NAMESPACE",
			Value: f.Namespace,
		},
```
Replace with:
```go
		{
			Name:  "FUNC_NAMESPACE",
			Value: f.Namespace,
		},
```

Find:
```go
			{
				Name:  "MOD_NAME",
				Value: "handler",
			},
			{
				Name:  "FUNC_HANDLER",
				Value: "main",
			},
```
Replace with:
```go
			{
				Name:  "HANDLER_MODULE_NAME",
				Value: "handler",
			},
			{
				Name:  "HANDLER_FUNCTION_NAME",
				Value: "main",
			},
```

- [ ] **Step 2: In `sourceEnvs()` (around line 698), rename one ENV var**

Find:
```go
		if f.HasPythonRuntime() {
			envs = append(envs, []corev1.EnvVar{
				{
					Name:  "FUNCTION_PATH",
					Value: "/kubeless",
				},
			}...)
		}
```
Replace with:
```go
		if f.HasPythonRuntime() {
			envs = append(envs, []corev1.EnvVar{
				{
					Name:  "HANDLER_FOLDER",
					Value: "/kubeless",
				},
			}...)
		}
```

- [ ] **Step 3: Run controller unit tests**

```bash
make -C components/buildless-serverless test
```

Expected: all tests pass. If any test asserts the old ENV name (e.g. `SERVICE_NAMESPACE`, `MOD_NAME`, `FUNC_HANDLER`, `FUNCTION_PATH`), update those assertions to the new names.

- [ ] **Step 4: Commit**

```bash
git add components/buildless-serverless/internal/controller/resources/deployment.go
git commit -m "feat(controller): rename injected ENVs to match new runtime naming convention"
```
