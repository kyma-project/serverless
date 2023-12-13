# Customize Function Traces

This tutorial shows how to use the built-in OpenTelemetry tracer object to send custom trace data to the trace backend.

Kyma Functions are instrumented to handle trace headers. This means that every time you call your Function, the executed logic is traceable using a dedicated span visible in the trace backend (that is, start time and duration).
Additionally, you can extend the default trace context and create your own custom spans as you wish (that is, when calling a remote service in your distributed application) or add additional information to the tracing context by introducing events and tags. The following tutorial shows you how to do it using tracer client that is available as part of the [event](../technical-reference/07-70-function-specification.md#event-object) object.

## Prerequisites

Before you start, make sure you have these tools installed:

- [Telemetry component installed](https://kyma-project.io/docs/kyma/latest/04-operation-guides/operations/02-install-kyma/#install-specific-components)
- [Trace pipeline configured](https://github.com/kyma-project/telemetry-manager/blob/main/docs/user/03-traces.md#setting-up-a-tracepipeline)

## Steps

The following code samples illustrate how to enrich the default trace with custom spans, events, and tags:

1. [Create an inline Function](01-10-create-inline-function.md) with the following body:

<!-- tabs:start -->

#### **Node.js**

      ```javascript

      const { SpanStatusCode } = require("@opentelemetry/api/build/src/trace/status");
      const axios = require("axios")
      module.exports = {
         main: async function (event, context) {

            const data = {
               name: "John",
               surname: "Doe",
               type: "Employee",
               id: "1234-5678"
            }

            const span = event.tracer.startSpan('call-to-acme-service');
            return await callAcme(data)
               .then(resp => {
                  if(resp.status!==200){
                    throw new Error("Unexpected response from acme service");
                  }
                  span.addEvent("Data sent");
                  span.setAttribute("data-type", data.type);
                  span.setAttribute("data-id", data.id);
                  span.setStatus({code: SpanStatusCode.OK});
                  return "Data sent";
               }).catch(err=> {
                  console.error(err)
                  span.setStatus({
                    code: SpanStatusCode.ERROR,
                    message: err.message,
                  });
                  return err.message;
               }).finally(()=>{
                  span.end();
               });
         }
      }

      let callAcme = (data)=>{
         return axios.post('https://acme.com/api/people', data)
      }
      ```

#### **Python**

      [OpenTelemetry SDK](https://opentelemetry.io/docs/instrumentation/python/manual/#traces) allows you to customize trace spans and events.
      Additionally, if you are using the `requests` library then all the HTTP communication can be auto-instrumented:

      ```python
      import requests
      import time
      from opentelemetry.instrumentation.requests import RequestsInstrumentor

      def main(event, context):
         # Create a new span to track some work
         with event.tracer.start_as_current_span("parent"):
            time.sleep(1)

            # Create a nested span to track nested work
            with event.tracer.start_as_current_span("child"):
               time.sleep(2)
               # the nested span is closed when it's out of scope

         # Now the parent span is the current span again
         time.sleep(1)

         # This span is also closed when it goes out of scope

         RequestsInstrumentor().instrument()

         # This request will be auto-intrumented
         r = requests.get('https://swapi.dev/api/people/2')
         return r.json()
      ```

<!-- tabs:end -->

2. [Expose your Function](01-20-expose-function.md).

3. Find the traces for the Function in the trace backend.