import * as opentelemetry from '@opentelemetry/api';
import { CompositePropagator, W3CTraceContextPropagator } from '@opentelemetry/core';
import { registerInstrumentations } from '@opentelemetry/instrumentation';
import { NodeTracerProvider, AlwaysOnSampler, ParentBasedSampler } from '@opentelemetry/sdk-trace-node';
import { SimpleSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { defaultResource, resourceFromAttributes } from '@opentelemetry/resources';
import { B3Propagator, B3InjectEncoding } from '@opentelemetry/propagator-b3';
import { ExpressInstrumentation, ExpressLayerType } from '@opentelemetry/instrumentation-express';
import { HttpInstrumentation } from '@opentelemetry/instrumentation-http';
import { ATTR_SERVICE_NAME } from '@opentelemetry/semantic-conventions';


const ignoredTargets = [
  "/healthz", "/favicon.ico", "/metrics"
]

export function setupTracer(functionName, traceCollectorEndpoint){

  const functionResource = resourceFromAttributes({
    [ATTR_SERVICE_NAME]: functionName,
  })

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

export function getCurrentSpan(){
  return opentelemetry.trace.getSpan(opentelemetry.context.active());
}

export function startNewSpan(name, tracer){
  const currentSpan = getCurrentSpan();
  const ctx = opentelemetry.trace.setSpan(
      opentelemetry.context.active(),
      currentSpan
  );
  return tracer.startSpan(name, undefined, ctx);
}
