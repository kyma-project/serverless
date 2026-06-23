import { configure as sdkConfigure } from 'sdk';
import { configureGracefulShutdown, handleTimeOut, isFunction, isPromise, handleError } from './lib/helper.js';
import bodyParser from 'body-parser';
import morgan from 'morgan';
import process from 'process';

import { setupTracer, getCurrentSpan } from './lib/tracer.js';
import { getMetrics, setupMetrics, createFunctionDurationHistogram, createFunctionCallsTotalCounter, createFunctionFailuresTotalCounter } from './lib/metrics.js';

// To catch unhandled exceptions thrown by user code async callbacks,
// these exceptions cannot be caught by try-catch in user function invocation code below
process.on('uncaughtException', (err) => {
    console.error(`Caught exception: ${err}`);
});

const handlerFolder = process.env.HANDLER_PATH || './function';
const handlerModuleName = process.env.HANDLER_MOD_NAME || 'handler';
const handlerFunctionName = process.env.HANDLER_FUNC_NAME || 'main';
const handlerPath = `${handlerFolder}/${handlerModuleName}.js`;

const serviceNamespace = process.env.SERVICE_NAMESPACE || '';
const functionName = process.env.FUNC_NAME || '';
const bodySizeLimit = Number(process.env.FUNC_BODY_MB_LIMIT || '1');
const serverHost = process.env.SERVER_HOST || '0.0.0.0';
const serverPort = Number(process.env.SERVER_PORT || '8080');
const timeout = Number(process.env.FUNC_TIMEOUT || '180');
const funcRuntime = process.env.FUNC_RUNTIME || 'nodejs26';
const traceCollectorEndpoint = process.env.TRACE_COLLECTOR_ENDPOINT || '';
const publisherProxyAddress = process.env.PUBLISHER_PROXY_ADDRESS || '';

console.log(`Importing function sources from ${handlerPath}:${handlerFunctionName}`);
console.log(`Tracing configured with endpoint ${traceCollectorEndpoint}`);
console.log(`Publisher Proxy available on address ${publisherProxyAddress}`);
console.log(`Starting ${funcRuntime} server ${serverHost}:${serverPort}`);

const tracer = setupTracer(functionName, traceCollectorEndpoint);
setupMetrics(functionName);
sdkConfigure(tracer, publisherProxyAddress, functionName, serviceNamespace, funcRuntime, timeout, bodySizeLimit);

const callsTotalCounter = createFunctionCallsTotalCounter(functionName);
const failuresTotalCounter = createFunctionFailuresTotalCounter(functionName);
const durationHistogram = createFunctionDurationHistogram(functionName);

//require express must be called AFTER tracer was setup!!!!!!
import express from 'express';
const app = express();

// User function.  Starts out undefined.
let userFunction;

app.use(bodyParser.json({ type: ['application/json', 'application/cloudevents+json'], limit: `${bodySizeLimit}mb`, strict: false }));
app.use(bodyParser.text({ type: ['text/*'], limit: `${bodySizeLimit}mb` }));
app.use(bodyParser.urlencoded({ limit: `${bodySizeLimit}mb`, extended: true }));
app.use(bodyParser.raw({ limit: `${bodySizeLimit}mb`, type: () => true }));

// Request logger
if (process.env['SERVER_INTERNAL_LOGGER']) {
    app.use(morgan('combined'));
}

app.use(handleTimeOut);

app.get('/healthz', (req, res) => res.status(200).send('OK'));
app.get('/metrics', (req, res) => getMetrics(req, res));
app.get('/favicon.ico', (req, res) => res.status(204).end());

// Generic route -- all http requests go to the user function.
// Since express 5.0.0 '*' wildcard requires a named variable (here: `path`)
// see https://github.com/pillarjs/path-to-regexp?tab=readme-ov-file#errors
app.all('*path', (req, res) => {
    res.header('Access-Control-Allow-Origin', '*');

    if (req.method === 'OPTIONS') {
        // CORS preflight support (Allow any method or header requested)
        res.header('Access-Control-Allow-Methods', req.headers['access-control-request-method']);
        res.header('Access-Control-Allow-Headers', req.headers['access-control-request-headers']);
        res.end();
        return;
    }

    callsTotalCounter.add(1);
    const startTime = new Date().getTime();

    if (!userFunction) {
        failuresTotalCounter.add(1);
        res.status(500).send('User function not loaded');
        return;
    }

    const currentSpan = getCurrentSpan();

    try {
        const out = userFunction(req, res);
        if (out && isPromise(out)) {
            out.catch((err) => {
                failuresTotalCounter.add(1);
                handleError(err, currentSpan, (body, status) => {
                    if (!res.writableEnded) res.status(status || 500).send(body);
                });
            });
        }
    } catch (err) {
        failuresTotalCounter.add(1);
        handleError(err, currentSpan, (body, status) => {
            if (!res.writableEnded) res.status(status || 500).send(body);
        });
    }

    const endTime = new Date().getTime();
    durationHistogram.record(endTime - startTime);
});

const server = app.listen(serverPort, serverHost);
configureGracefulShutdown(server);

const startTime = process.hrtime();
import(handlerPath).then((fn) => {
    if (isFunction(fn[handlerFunctionName])) {
        userFunction = fn[handlerFunctionName];
        const elapsed = process.hrtime(startTime);
        console.log(`user code loaded in ${elapsed[0]}sec ${elapsed[1] / 1000000}ms`);
    } else {
        console.error(`Content loaded from ${handlerPath} is not a function. Make sure your function exports '${handlerFunctionName}' function`, fn);
    }
}).catch((err) => {
    console.error('Failed to load user function:', err);
});
