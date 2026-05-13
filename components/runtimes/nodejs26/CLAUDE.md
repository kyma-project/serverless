# Node.js 26 Runtime

Express 5 server running on Node.js 25+ (placeholder until Node 26 LTS is released).

## Handler API

User functions receive raw Express objects:
```js
module.exports = {
    main: function (req, res) {
        res.send('Hello!');
    }
}
```

## File Layout

- `server.mjs` — Entry point. Reads all env vars, sets up tracer/metrics, loads user handler via dynamic `import(handlerPath)`
- `sdk/index.mjs` — ESM package exposing `getCloudEvent(req)`, `emitCloudEvent()`, `getTracer()`, metadata getters. Registered as local dep in package.json (`"sdk": "file:./sdk"`) so users can `import { ... } from 'sdk'`
- `lib/tracer.js` — OpenTelemetry tracer setup (NodeTracerProvider, OTLP exporter, Express+HTTP instrumentation)
- `lib/metrics.js` — Prometheus metrics via OpenTelemetry SDK (PrometheusExporter)
- `lib/helper.js` — Request timeout middleware, graceful shutdown, error handling utilities (CommonJS)

## Key Design Decisions

- **ESM throughout**: `server.mjs` and `sdk/index.mjs` are ESM. `lib/*.js` files remain CommonJS (they predate the rework and are internal).
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
HANDLER_PATH=./handler.js npm start
```
