# Python 3.14 Runtime

Flask + gevent WSGI server running on Python 3.14.

## Handler API

User functions take no arguments — use Flask's request context-local:

```python
import flask

def main():
    name = flask.request.args.get('name', 'World')
    return f'Hello {name}!'
```

## Handler Loading and `HANDLER_PATH`

`server.py` resolves the handler via two env vars:
- `HANDLER_PATH` (default: `/`) — appended to `sys.path`; `importlib.import_module(HANDLER_MOD_NAME)` finds `handler.py` there.
- `FUNCTION_PATH` (default: `/`) — also appended to `sys.path` for `kyma cli function eject` compatibility.

Two layouts are supported:

- **`kyma cli function eject`**: `handler.py` sits flat next to `server.py`; default `HANDLER_PATH=/` and `FUNCTION_PATH=/` resolve it from the working directory.
- **buildless-serverless**: the controller runs `/usr/src/app/start.sh`, which writes sources to `/usr/src/app/function/` and sets `FUNCTION_PATH=/usr/src/app/function` in the Pod so Python can find `handler.py` there.

## File Layout

- `server.py` — Entry point. Reads env vars, configures sdk/tracer, sets up Flask app with routes, wraps with WSGILogger if logging enabled, runs gevent WSGIServer with graceful shutdown
- `lib/sdk.py` — User-facing SDK: `get_cloud_event()`, `emit_cloud_event()`, `get_tracer()`, metadata getters. Uses `flask.request` internally for CloudEvent parsing
- `lib/tracing.py` — OpenTelemetry tracer setup (OTLP exporter, B3 propagation, requests auto-instrumentation)
- `lib/module.py` — `Handler` class: imports user module, wraps calls with Prometheus metrics and gevent.Timeout

## Key Design Decisions

- **`sys.path.insert(0, 'lib/')` before all imports** (line 7 of server.py): Ensures `import sdk` in both server.py and user handlers resolves to the same `sys.modules['sdk']` entry. Without this, server would register `lib.sdk` and user code would get a separate unconfigured `sdk` module.
- **No arguments to handler**: Flask's thread-local `flask.request` provides per-request isolation natively via gevent greenlets. No need to pass request explicitly.
- **gevent.Timeout for FUNC_TIMEOUT**: User function calls are wrapped in `gevent.Timeout(seconds)` — returns 408 if exceeded (matches python312 behavior).
- **Graceful shutdown**: SIGTERM/SIGINT handlers call `server.stop()` via `gevent.signal_handler`.
- **Request body limit**: `app.config['MAX_CONTENT_LENGTH'] = func_body_mb_limit * 1024 * 1024` (Flask built-in, from `FUNC_BODY_MB_LIMIT` env var in MB).

## Dependencies

`requirements.txt`: flask, gevent, wsgi-request-logger, prometheus_client, opentelemetry-*, cloudevents, requests.

## Running

```bash
make run                    # creates venv, pip install, python server.py
# or manually:
python3 -m venv .venv
.venv/bin/pip install -r requirements.txt
echo 'def main(): return "ok"' > handler.py
.venv/bin/python server.py
```
