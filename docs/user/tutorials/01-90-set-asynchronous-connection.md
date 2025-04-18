# Set Asynchronous Communication Between Functions

This tutorial demonstrates how to connect two Functions asynchronously. It is based on the [in-cluster Eventing example](https://github.com/kyma-project/serverless/tree/main/examples/incluster_eventing).

The example provides a very simple scenario of asynchronous communication between two Functions. The first Function accepts the incoming traffic via HTTP, sanitizes the payload, and publishes the content as an in-cluster event using [Kyma Eventing](https://kyma-project.io/docs/kyma/latest/01-overview/eventing/).
The second Function is a message receiver. It subscribes to the given event type and stores the payload.

This tutorial shows only one possible use case. There are many more use cases on how to orchestrate your application logic into specialized Functions and benefit from decoupled, re-usable components and event-driven architecture.

## Prerequisites

- [Kyma CLI](https://github.com/kyma-project/cli)
- [Eventing, Istio, and API-Gateway modules added](https://kyma-project.io/#/02-get-started/01-quick-install)
- [The cluster domain set up](https://kyma-project.io/#/api-gateway/user/tutorials/01-10-setup-custom-domain-for-workload)
  
## Steps

1. Export the `KUBECONFIG` variable:

   ```bash
   export KUBECONFIG={KUBECONFIG_PATH}
   ```

2. Create the `emitter` and `receiver` folders in your project.

### Create the Emitter Function

1. Go to the `emitter` folder and run Kyma CLI `init` command to initialize the scaffold for your first Function:

   ```bash
   kyma alpha function init
   ```

   The `init` command creates these files in your workspace folder:

   - `handler.js` with the Function's code and the simple "Hello Serverless" logic
  
   - `package.json` with the Function's dependencies

2. Provide your Function logic in the `handler.js` file:

   > [!NOTE]
   > In this example, there's no sanitization logic. The `sanitize` Function is just a placeholder.

   ```js
   module.exports = {
      main: async function (event, context) {
         let sanitisedData = sanitise(event.data)

         const eventType = "sap.kyma.custom.acme.payload.sanitised.v1";
         const eventSource = "kyma";
         
         return await event.emitCloudEvent(eventType, eventSource, sanitisedData)
               .then(resp => {
                  return "Event sent";
               }).catch(err=> {
                  console.error(err)
                  return err;
               });
      }
   }
   let sanitise = (data)=>{
      console.log(`sanitising data...`)
      console.log(data)
      return data
   }
   ```

   The `sap.kyma.custom.acme.payload.sanitised.v1` is a sample event type that the emitter Function declares when publishing events. You can choose a different one that better suits your use case. Keep in mind the constraints described on the [Event names](https://kyma-project.io/docs/kyma/latest/05-technical-reference/evnt-01-event-names/) page. The receiver subscribes to the event type to consume the events.

   The event object provides convenience functions to build and publish events. To send the event, build the Cloud Event. To learn more, read [Function's specification](../technical-reference/07-70-function-specification.md#event-object-sdk). In addition, your **eventOut.source** key must point to `“kyma”` to use Kyma in-cluster Eventing.
   There is a `require('axios')` line even though the Function code is not using it directly. This is needed for the auto-instrumentation to properly handle the outgoing requests sent using the `publishCloudEvent` method (which uses `axios` library under the hood). Without the `axios` import the Function still works, but the published events are not reflected in the trace backend.

3. Apply your emitter Function:

   ```bash
   kyma alpha function create emitter --source handler.js --dependencies package.json
   ```

   Your Function is now built and deployed in Kyma runtime. Kyma exposes it through the APIRule. The incoming payloads are processed by your emitter Function. It then sends the sanitized content to the workload that subscribes to the selected event type. In our case, it's the receiver Function.

4. Expose Function by creating the APIRule CR:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: gateway.kyma-project.io/v2
   kind: APIRule
   metadata:
     name: incoming-http-trigger
   spec:
     hosts:
     - incoming
     service:
       name: emitter
       namespace: default
       port: 80
     gateway: kyma-system/kyma-gateway
     rules:
     - path: /*
       methods: ["GET", "POST"]
       noAuth: true
   EOF
   ```

5. Test the first Function. Send the payload and see if your HTTP traffic is accepted:

   ```bash
   export KYMA_DOMAIN={KYMA_DOMAIN_VARIABLE}
   curl -X POST "https://incoming.${KYMA_DOMAIN}" -H 'Content-Type: application/json' -d '{"foo":"bar"}'
   ```

### Create the Receiver Function

1. Go to your `receiver` folder and run Kyma CLI `init` command to initialize the scaffold for your second Function:

   ```bash
   kyma alpha function init
   ```

   The `init` command creates the same files as in the `emitter` folder.

3. Apply your receiver Function:

   ```bash
   kyma alpha function create receiver --source handler.js --dependencies package.json
   ```

   The Function is configured, built, and deployed in Kyma runtime. The Subscription becomes active and all events with the selected type are processed by the Function.  

2. Subscribe the `receiver` Function to the event:  

   ```bash
   cat <<EOF | kubectl apply -f -
      apiVersion: eventing.kyma-project.io/v1alpha2
      kind: Subscription
      metadata:
         name: event-receiver
         namespace: default
      spec:
         sink: 'http://receiver.default.svc.cluster.local'
         source: ""
         types:
         - sap.kyma.custom.acme.payload.sanitised.v1
   EOF
   ```

### Test the Whole Setup

Send a payload to the first Function. For example, use the POST request mentioned above. As the Functions are joined by the in-cluster Eventing, the payload is processed in sequence by both of your Functions.
In the Function's logs, you can see that both sanitization logic (using the first Function) and the storing logic (using the second Function) are executed.
