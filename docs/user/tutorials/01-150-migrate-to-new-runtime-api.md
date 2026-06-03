# Migrate Functions to the New Runtime API

This tutorial shows how to migrate an existing Function from `nodejs22`, `nodejs24`, or `python312` to the new `nodejs26` or `python314` runtime.

## Prerequisites

- You have a Function deployed with runtime `nodejs22`, `nodejs24`, or `python312`.

## Context

The `nodejs26` and `python314` runtimes introduce a new handler API that removes the `event` and `context` wrapper objects. Handlers receive the raw HTTP framework objects directly (`req`/`res` in Node.js, `flask.request` in Python), and function metadata previously available through `context` is now provided by an explicit `sdk` module. Migrating gives you direct access to the underlying framework, a typed CloudEvent object, and a consistent SDK across both runtimes.

## Procedure

### 1. Update the Handler signature

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

### 2. Update SDK Usage

Previously, function metadata was available through the `context` object passed to the handler. The `sdk` module replaces it with explicit functions. For more information, see [Function's Specification](../technical-reference/07-70-function-specification.md#sdk-module).

<!-- tabs:start -->

#### **Node.js**

```javascript
// Before — context fields and event.tracer
module.exports = {
    main: async function (event, context) {
        const tracer = event.tracer;
        console.log(context['function-name']);
        console.log(context['namespace']);
        console.log(context['runtime']);
        console.log(context['timeout']);
        console.log(context['body-size-limit']);
    }
}
```

```javascript
// After — sdk functions
import { getTracer, getFunctionName, getNamespace, getRuntime, getTimeout, getBodySizeLimit } from 'sdk';

export function main(req, res) {
    const tracer = getTracer();
    console.log(getFunctionName());
    console.log(getNamespace());
    console.log(getRuntime());
    console.log(getTimeout());
    console.log(getBodySizeLimit());
    res.send('ok');
}
```

#### **Python**

```python
# Before — context fields and event.tracer
def main(event, context):
    tracer = event.tracer
    print(context['function-name'])
    print(context['namespace'])
    print(context['runtime'])
    print(context['timeout'])
    print(context['body-size-limit'])
    return "ok"
```

```python
# After — sdk functions
import sdk

def main():
    tracer = sdk.get_tracer()
    print(sdk.get_function_name())
    print(sdk.get_namespace())
    print(sdk.get_runtime())
    print(sdk.get_timeout())
    print(sdk.get_body_size_limit())
    return "ok"
```

<!-- tabs:end -->

### 3. Update CloudEvent Handling

Previously, CloudEvent attributes were spread as individual fields of the `event` object. The `sdk` module returns a standard `CloudEvent` object with typed properties.

<!-- tabs:start -->

#### **Node.js**

```javascript
// Before — CloudEvent fields were individual keys on the event object
module.exports = {
    main: async function (event, context) {
        if (event['ce-type']) {
            console.log(event['ce-type']);
            console.log(event['ce-source']);
            // eventtypeversion was dropped in CloudEvents version 0.2
            // console.log(event['ca-eventtypeversion']);
            console.log(event['ce-specversion']);
            console.log(event['ce-id']);
            console.log(event['ce-time']);
            console.log(event['ce-datacontenttype']);
            console.log(event['data']);
            return event['ce-type'];
        }
        return "not a cloud event";
    }
}
```

```javascript
// After — getCloudEvent() returns a standard CloudEvent object
import { getCloudEvent } from 'sdk';

export function main(req, res) {
    const ce = getCloudEvent(req);
    if (ce) {
        console.log(ce.type);
        console.log(ce.source);
        console.log(ce.specversion);
        console.log(ce.id);
        console.log(ce.time);
        console.log(ce.datacontenttype);
        console.log(ce.data);
        res.send(ce.type);
        return;
    }
    res.send("not a cloud event");
}
```

#### **Python**

```python
# Before — CloudEvent fields were individual keys on the event object
def main(event, context):
    if event['ce-type']:
        print(event['ce-type'])
        print(event['ce-source'])
        # eventtypeversion was dropped in CloudEvents version 0.2
        # print(event['ce-eventtypeversion'])
        print(event['ce-specversion'])
        print(event['ce-id'])
        print(event['ce-time'])
        print(event['ce-datacontenttype'])
        print(event['data'])
        return event['ce-type']
    return "not a cloud event"
```

```python
# After — get_cloud_event() returns a standard CloudEvent object
import sdk

def main():
    ce = sdk.get_cloud_event()
    if ce:
        print(ce.get_type())
        print(ce.get_source())
        print(ce.get_specversion())
        print(ce.get_id())
        print(ce.get_time())
        print(ce.get_datacontenttype())
        print(ce.get_data())
        return ce.get_type()
    return "not a cloud event"
```

<!-- tabs:end -->

### 4. Emit CloudEvents

Previously, CloudEvents were emitted using `event.emitCloudEvent()`. To emit a CloudEvent in the new runtimes, import `emitCloudEvent` from the `sdk` module and call it directly.

<!-- tabs:start -->

#### **Node.js**

```javascript
// Before
module.exports = {
    main: async function (event, context) {
        await event.emitCloudEvent(
            'com.example.order.created',
            '/orders',
            { orderId: '123' }
        );
        return "event emitted";
    }
}
```

```javascript
// After
import { emitCloudEvent } from 'sdk';

export async function main(req, res) {
    await emitCloudEvent(
        'com.example.order.created',
        '/orders',
        { orderId: '123' }
    );
    res.send("event emitted");
}
```

#### **Python**

```python
# Before
def main(event, context):
    event.emitCloudEvent(
        'com.example.order.created',
        '/orders',
        {'orderId': '123'}
    )
    return "event emitted"
```

```python
# After
import sdk

def main():
    sdk.emit_cloud_event(
        'com.example.order.created',
        '/orders',
        {'orderId': '123'}
    )
    return "event emitted"
```

> **Note:** In `python314`, `datacontenttype` is required in `optional_attributes`. In `nodejs26`, it is inferred automatically if omitted.

<!-- tabs:end -->

### 5. Update HTTP Responses

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

Return values work the same way as in the legacy runtime – Flask response tuples are supported.

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

### 6. Update Environment Variables

If you override any of the following environment variables in your Function CR, update them to the new names and values:

| Old name                       | New name                 | Runtimes            | Notes                                |
| ------------------------------ | ------------------------ | ------------------- | ------------------------------------ |
| `FUNC_HANDLER`                 | `HANDLER_FUNC_NAME`      | nodejs26, python314 |                                      |
| `MOD_NAME`                     | `HANDLER_MOD_NAME`       | nodejs26, python314 |                                      |
| `KUBELESS_INSTALL_VOLUME`      | `HANDLER_PATH`           | nodejs26, python314 |                                      |
| `REQ_MB_LIMIT`                 | `FUNC_BODY_MB_LIMIT`     | nodejs26            |                                      |
| `FUNC_MEMFILE_MAX`             | `FUNC_BODY_MB_LIMIT`     | python314           | Unit changed from bytes to megabytes |
| `KYMA_INTERNAL_LOGGER_ENABLED` | `SERVER_INTERNAL_LOGGER` | nodejs26, python314 |                                      |

### 7. Change the Runtime Version

Update `spec.runtime` in your Function CR to the new runtime:

<!-- tabs:start -->

#### **Node.js**

```bash
kubectl patch function <function-name> --type merge -p '{"spec":{"runtime":"nodejs26"}}'
```

#### **Python**

```bash
kubectl patch function <function-name> --type merge -p '{"spec":{"runtime":"python314"}}'
```

<!-- tabs:end -->

## Result

Verify that the Function is running with the new runtime:

```bash
kubectl get function <function-name>
```

The `STATE` column shows `Running` when the migration is complete.

## Related Information

- [Function's Specification](../technical-reference/07-70-function-specification.md) — full SDK reference for `nodejs26` and `python314`
- [Environment Variables](../technical-reference/05-20-env-variables.md) — complete list of runtime environment variables
