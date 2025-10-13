const opentelemetry = require('@opentelemetry/api');
const { MeterProvider } = require('@opentelemetry/sdk-metrics');
const { PrometheusExporter } = require('@opentelemetry/exporter-prometheus');
const { defaultResource, resourceFromAttributes } = require( '@opentelemetry/resources');
const { ATTR_SERVICE_NAME } = require('@opentelemetry/semantic-conventions');


let exporter;

function setupMetrics(functionName){

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

function createFunctionCallsTotalCounter(name){
  const meter =  opentelemetry.metrics.getMeter(name)
  return meter.createCounter('function_calls_total',{
    description: 'Number of calls to user function',
  }); 
}
  
  
function createFunctionFailuresTotalCounter(name){
  const meter =  opentelemetry.metrics.getMeter(name)
  return meter.createCounter('function_failures_total',{
    description: 'Number of exceptions in user function',
  });  
}

function createFunctionDurationHistogram(name){
  const meter =  opentelemetry.metrics.getMeter(name)
  return meter.createHistogram("function_duration_miliseconds",{
    description: 'Duration of user function in miliseconds',
  });  
}

const getMetrics = (req, res) => {
  exporter.getMetricsRequestHandler(req, res);
};

module.exports = {
    setupMetrics,
    createFunctionCallsTotalCounter,
    createFunctionFailuresTotalCounter,
    createFunctionDurationHistogram,
    getMetrics,
}