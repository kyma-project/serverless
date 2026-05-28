# Migrate Functions to the New Runtime API

The `nodejs26` and `python314` runtimes introduce a new handler API that replaces the legacy `event` and `context` arguments. This tutorial shows how to migrate existing Functions from `nodejs22`, `nodejs24`, or `python312` to the new runtimes.

## Handler Signature

<!-- tabs:start -->

#### **Node.js**

The handler now receives raw Express `req` and `res` objects instead of the custom `event` and `context`.

| Before (nodejs22/nodejs24) | After (nodejs26) |
| -------------------------- | ---------------- |
| `main(event, context)`     | `main(req, res)` |

```javascript
// Before
module.exports = {
    main: function (event, context) {
        return "Hello, " + event.data.name;
    }
}
```

```javascript
// After
module.exports = {
    main: function (req, res) {
        res.send("Hello, " + req.body.name);
    }
}
```

#### **Python**

The handler now takes no arguments. To access the incoming request, use `flask.request`.

| Before (python312)     | After (python314) |
| ---------------------- | ----------------- |
| `main(event, context)` | `main()`          |

```python
# Before
def main(event, context):
    return "Hello, " + event.data["name"]
```

```python
# After
import flask

def main():
    data = flask.request.get_json()
    return "Hello, " + data["name"]
```

<!-- tabs:end -->

## SDK Functions

The `sdk` module replaces the helpers that were previously embedded in the `event` object. To use it, import it in your handler. [Function's specification](../technical-reference/07-70-function-specification.md) contains a with list of all available functions.

Examples:
<!-- tabs:start -->

#### **Node.js**

```javascript
// Before
module.exports = {
    main: async function (event, context) {
        const ce = event.ce;
        const tracer = event.tracer;
        console.log(context.funcName);
    }
}
```

```javascript
// After
import { getCloudEvent, getTracer, getFunctionName } from 'sdk';

export function main(req, res) {
    const ce = getCloudEvent(req);
    const tracer = getTracer();
    console.log(getFunctionName());
    res.send('ok');
}
```

#### **Python**

```python
# Before
def main(event, context):
    ce = event.ce
    tracer = event.tracer
    print(context.function_name)
    return "ok"
```

```python
# After
import sdk

def main():
    ce = sdk.get_cloud_event()
    tracer = sdk.get_tracer()
    print(sdk.get_function_name())
    return "ok"
```

<!-- tabs:end -->

## CloudEvents

<!-- tabs:start -->

#### **Node.js**

```javascript
// Before
module.exports = {
    main: async function (event, context) {
        if (event.ce) {
            return event.ce.type;
        }
        return "not a cloud event";
    }
}
```

```javascript
// After
import { getCloudEvent } from 'sdk';

export function main(req, res) {
    const ce = getCloudEvent(req);
    if (ce) {
        res.send(ce.type);
        return;
    }
    res.send("not a cloud event");
}
```

#### **Python**

```python
# Before
def main(event, context):
    if event.ce:
        return event.ce["type"]
    return "not a cloud event"
```

```python
# After
import sdk

def main():
    ce = sdk.get_cloud_event()
    if ce:
        return ce.get_type()
    return "not a cloud event"
```

<!-- tabs:end -->

## HTTP Responses

<!-- tabs:start -->

#### **Node.js**

To send a response, use the Express `res` object directly instead of returning a value.

```javascript
// Before
module.exports = {
    main: function (event, context) {
        return { statusCode: 400, body: "Bad request" };
    }
}
```

```javascript
// After
module.exports = {
    main: function (req, res) {
        res.status(400).send("Bad request");
    }
}
```

#### **Python**

Return values work the same way as in the legacy runtime — Flask response tuples are supported.

```python
# Before
def main(event, context):
    return "Bad request", 400
```

```python
# After — no change needed
def main():
    return "Bad request", 400
```

<!-- tabs:end -->

## Environment Variables

The following environment variables were renamed in the new runtimes:

| Old name                       | New name                 | Runtimes            |
| ------------------------------ | ------------------------ | ------------------- |
| `FUNC_HANDLER`                 | `HANDLER_FUNC_NAME`      | nodejs26, python314 |
| `MOD_NAME`                     | `HANDLER_MOD_NAME`       | nodejs26, python314 |
| `KUBELESS_INSTALL_VOLUME`      | `HANDLER_PATH`           | nodejs26, python314 |
| `REQ_MB_LIMIT`                 | `FUNC_BODY_MB_LIMIT`     | nodejs26            |
| `FUNC_MEMFILE_MAX`             | `FUNC_BODY_MB_LIMIT`     | python314           |
| `KYMA_INTERNAL_LOGGER_ENABLED` | `SERVER_INTERNAL_LOGGER` | nodejs26, python314 |

Update any environment variable overrides in your Function CR accordingly.
