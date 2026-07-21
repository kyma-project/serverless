# Node.js 26 Runtime

Express 5 server running on Node.js 26.

## Handler API

User functions receive raw Express objects. Handler files default to **CommonJS** (no `"type": "module"` in package.json):

```js
const { getCloudEvent } = require('sdk');

module.exports = {
    main: function (req, res) {
        res.send('Hello!');
    }
}
```

Use `.mjs` extension or add `"type": "module"` to your handler's package.json to write ESM (`import`/`export`) instead.

## Handler Loading and `HANDLER_PATH`

`server.mjs` builds the handler path with `path.join(HANDLER_PATH, HANDLER_MOD_NAME + '.js')` (defaults: `HANDLER_PATH=./`, `HANDLER_MOD_NAME=handler` → `./handler.js`). A leading `./` is re-prefixed for relative results so dynamic `import()` does not treat the path as a bare specifier.

Two layouts are supported:

- **`kyma cli function eject`**: handler files sit flat next to `server.mjs`; default `HANDLER_PATH=./` works as-is.
- **buildless-serverless**: the controller mounts sources at `/usr/src/app/function/` and sets `HANDLER_PATH=./function` in the Pod, pointing the server at the mounted directory. The container runs `/usr/src/app/start.sh` (absolute path, because `workingDir` is the sources mount, not `/usr/src/app`).

## File Layout

- `server.mjs` — Entry point. Reads all env vars, sets up tracer/metrics, loads user handler via dynamic `import(handlerPath)`
- `start.sh` — Container startup script. Writes the inline `FUNC_HANDLER_SOURCE`/`FUNC_HANDLER_DEPENDENCIES` envs (or copies git sources) into the working directory, runs `npm install`, then `cd ..` and `npm start`
- `sdk/index.js` — CJS package exposing `getCloudEvent(req)`, `emitCloudEvent()`, `getTracer()`, metadata getters. Registered as a local dep in package.json (`"sdk": "file:./sdk"`); user handlers consume it via `const { ... } = require('sdk')`
- `lib/tracer.js` — OpenTelemetry tracer setup (NodeTracerProvider, OTLP exporter, Express+HTTP instrumentation)
- `lib/metrics.js` — Prometheus metrics via OpenTelemetry SDK (PrometheusExporter)
- `lib/helper.js` — Request timeout middleware, graceful shutdown, error handling utilities

## Key Design Decisions

- **ESM entry point, CJS internals**: `server.mjs` is ESM. `sdk/index.js` and `lib/*.js` are CommonJS — Node.js ESM can import them via named export detection. Converting lib/sdk to ESM is deferred to a future PR.
- **Express must be imported after tracer setup**: OpenTelemetry HTTP/Express instrumentation patches modules at require-time. The `import express` is hoisted by ESM, but `app.listen()` must come after `setupTracer()`.
- **User function loaded late**: `import(handlerPath)` runs after the server is already listening. If loading fails, healthz still responds but requests get 500.
- **No sendResponse abstraction**: Unlike legacy runtimes, the user controls `res` directly. The server only catches unhandled promise rejections/throws as a safety net.

## Dependencies

Key runtime deps: `express@5`, `body-parser`, `morgan`, `cloudevents`, `axios`, `@opentelemetry/*`.

The `sdk` package is a `file:` dependency — `npm install` symlinks `node_modules/sdk` → `./sdk/`.

## Running

```bash
make run                    # npm install + npm start
# or manually:
npm install
echo 'module.exports = { main: (req, res) => res.send("ok") }' > handler.js
npm start
```
