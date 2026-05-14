import * as opentelemetry from '@opentelemetry/api';
import { MeterProvider } from '@opentelemetry/sdk-metrics';
import { PrometheusExporter } from '@opentelemetry/exporter-prometheus';
import { defaultResource, resourceFromAttributes } from '@opentelemetry/resources';
import { ATTR_SERVICE_NAME } from '@opentelemetry/semantic-conventions';


let exporter;

export function setupMetrics(functionName){

    exporter = new PrometheusExporter(
        { preventServerStart: true},
    );

    const functionResource = resourceFromAttributes({
      [ATTR_SERVICE_NAME]: functionName,
    });

    const myServiceMeterProvider = new MeterProvider({
      resource: functionResource.merge(defaultResource()),
      readers: [exporter],
    });

    opentelemetry.metrics.setGlobalMeterProvider(myServiceMeterProvider);

}

export function createFunctionCallsTotalCounter(name){
  const meter = opentelemetry.metrics.getMeter(name)
  return meter.createCounter('function_calls_total',{
    description: 'Number of calls to user function',
  });
}

export function createFunctionFailuresTotalCounter(name){
  const meter = opentelemetry.metrics.getMeter(name)
  return meter.createCounter('function_failures_total',{
    description: 'Number of exceptions in user function',
  });
}

export function createFunctionDurationHistogram(name){
  const meter = opentelemetry.metrics.getMeter(name)
  return meter.createHistogram("function_duration_miliseconds",{
    description: 'Duration of user function in miliseconds',
  });
}

export const getMetrics = (req, res) => {
  exporter.getMetricsRequestHandler(req, res);
};
