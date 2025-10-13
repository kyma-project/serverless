'use strict';

const opentelemetry = require('@opentelemetry/api');
const { CompositePropagator, W3CTraceContextPropagator } = require( '@opentelemetry/core');
const { registerInstrumentations } = require( '@opentelemetry/instrumentation');
const { NodeTracerProvider, AlwaysOnSampler, ParentBasedSampler } = require( '@opentelemetry/sdk-trace-node');
const { SimpleSpanProcessor } = require( '@opentelemetry/sdk-trace-base');
const { OTLPTraceExporter } =  require('@opentelemetry/exporter-trace-otlp-http');
const { defaultResource, resourceFromAttributes } = require( '@opentelemetry/resources');
const { B3Propagator, B3InjectEncoding } = require("@opentelemetry/propagator-b3");
const { ExpressInstrumentation, ExpressLayerType } = require( '@opentelemetry/instrumentation-express');
const { HttpInstrumentation } = require('@opentelemetry/instrumentation-http');
const { ATTR_SERVICE_NAME } = require('@opentelemetry/semantic-conventions');
const axios = require("axios")


const ignoredTargets = [
  "/healthz", "/favicon.ico", "/metrics"
]

function setupTracer(functionName){
  
  const functionResource = resourceFromAttributes({
    [ATTR_SERVICE_NAME]: functionName,
  })

  const traceCollectorEndpoint = process.env.TRACE_COLLECTOR_ENDPOINT;

  let spanProcessors = [];

  if(traceCollectorEndpoint){
    const exporter = new OTLPTraceExporter({
      url: traceCollectorEndpoint
    });
    spanProcessors.push(new SimpleSpanProcessor(exporter));
  }

  const provider = new NodeTracerProvider({
    resource: functionResource.merge(defaultResource()),
    sampler: new ParentBasedSampler({
      root: new AlwaysOnSampler(),
    }),
    spanProcessors,
  });

  const propagator = new CompositePropagator({
    propagators: [
      new W3CTraceContextPropagator(), 
      new B3Propagator({injectEncoding: B3InjectEncoding.MULTI_HEADER})
    ],
  })

  registerInstrumentations({
    tracerProvider: provider,
    instrumentations: [
      new HttpInstrumentation({
        ignoreIncomingRequestHook: (req) => {
          // Ignore requests to healthz, favicon.ico and metrics endpoints
          return ignoredTargets.includes(req.url);
        }
      }),
      new ExpressInstrumentation({
        ignoreLayersType: [ExpressLayerType.MIDDLEWARE],
      }),
    ],
  });


  // Initialize the OpenTelemetry APIs to use the NodeTracerProvider bindings
  provider.register({
    propagator: propagator,
  });

  return opentelemetry.trace.getTracer("io.kyma-project.serverless");
};

module.exports = {
    setupTracer,
    startNewSpan,
    getCurrentSpan,
}


function getCurrentSpan(){
  return opentelemetry.trace.getSpan(opentelemetry.context.active());
}

function startNewSpan(name, tracer){
  const currentSpan = getCurrentSpan();
  const ctx = opentelemetry.trace.setSpan(
      opentelemetry.context.active(),
      currentSpan
  );
  return tracer.startSpan(name, undefined, ctx);
}
