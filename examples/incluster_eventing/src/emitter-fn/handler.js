const { SpanStatusCode } = require("@opentelemetry/api");

module.exports = {
    main: async function (event, context) {
        let sanitisedData = sanitise(event.data)

        const eventType = process.env['EVENT_TYPE'];
        const eventSource = process.env['EVENT_SOURCE'];

        const span = event.tracer.startSpan('call-to-kyma-eventing');
        
        // you can pass additional cloudevents attributes  
        // const eventtypeversion = "v1";
        // const datacontenttype = "application/json";
        // return await event.emitCloudEvent(eventType, eventSource, sanitisedData, {eventtypeversion, datacontenttype})
        
        return await event.emitCloudEvent(eventType, eventSource, sanitisedData)
            .then(resp => {
                console.log(resp.status);
                span.addEvent("Event sent");
                span.setAttribute("event-type", eventType);
                span.setAttribute("event-source", eventSource);
                span.setStatus({code: SpanStatusCode.OK});
                return "Event sent";
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
let sanitise = (data)=>{
    console.log(`sanitising data...`)
    console.log(data)
    return data
}