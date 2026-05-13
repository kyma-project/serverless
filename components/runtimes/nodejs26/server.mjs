import sdk from './lib/sdk.js';
import helper from './lib/helper.js';
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

const funcName = process.env.FUNC_NAME || '';
const funcNamespace = process.env.SERVICE_NAMESPACE || '';
const funcRuntime = process.env.FUNC_RUNTIME || 'nodejs26';
const serverHost = process.env.SERVER_HOST || '0.0.0.0';
const serverPort = parseInt(process.env.SERVER_PORT || '8080', 10);
const serverCallTimeout = Number(process.env.FUNC_TIMEOUT || '180');
const reqMbLimit = Number(process.env.REQ_MB_LIMIT || '1');
const handlerPath = process.env.HANDLER_PATH || './handler.js';
const traceCollectorEndpoint = process.env.TRACE_COLLECTOR_ENDPOINT || '';
const publisherProxyAddress = process.env.PUBLISHER_PROXY_ADDRESS || '';

console.log(`Importing function sources from ${handlerPath}:main`);
console.log(`Tracing configured with endpoint ${traceCollectorEndpoint}`);
console.log(`Publisher Proxy available on address ${publisherProxyAddress}`);
console.log(`Starting ${funcRuntime} server ${serverHost}:${serverPort}`);

const tracer = setupTracer(funcName);
setupMetrics(funcName);
sdk._configure(tracer, publisherProxyAddress, funcName, funcNamespace, funcRuntime, serverCallTimeout, reqMbLimit);

const callsTotalCounter = createFunctionCallsTotalCounter(funcName);
const failuresTotalCounter = createFunctionFailuresTotalCounter(funcName);
const durationHistogram = createFunctionDurationHistogram(funcName);

//require express must be called AFTER tracer was setup!!!!!!
import express from 'express';
const app = express();

// User function.  Starts out undefined.
let userFunction;

app.use(bodyParser.json({ type: ['application/json', 'application/cloudevents+json'], limit: `${reqMbLimit}mb`, strict: false }));
app.use(bodyParser.text({ type: ['text/*'], limit: `${reqMbLimit}mb` }));
app.use(bodyParser.urlencoded({ limit: `${reqMbLimit}mb`, extended: true }));
app.use(bodyParser.raw({ limit: `${reqMbLimit}mb`, type: () => true }));

// Request logger
if (process.env['KYMA_INTERNAL_LOGGER_ENABLED']) {
    app.use(morgan('combined'));
}

app.use(helper.handleTimeOut);

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
        if (out && helper.isPromise(out)) {
            out.catch((err) => {
                failuresTotalCounter.add(1);
                helper.handleError(err, currentSpan, (body, status) => {
                    if (!res.writableEnded) res.status(status || 500).send(body);
                });
            });
        }
    } catch (err) {
        failuresTotalCounter.add(1);
        helper.handleError(err, currentSpan, (body, status) => {
            if (!res.writableEnded) res.status(status || 500).send(body);
        });
    }

    const endTime = new Date().getTime();
    durationHistogram.record(endTime - startTime);
});

const server = app.listen(serverPort, serverHost);
helper.configureGracefulShutdown(server);

const startTime = process.hrtime();
import(handlerPath).then((fn) => {
    if (helper.isFunction(fn.main)) {
        userFunction = fn.main;
        const elapsed = process.hrtime(startTime);
        console.log(`user code loaded in ${elapsed[0]}sec ${elapsed[1] / 1000000}ms`);
    } else {
        console.error("Content loaded is not a function. Make sure your function exports 'main' function", fn);
    }
}).catch((err) => {
    console.error('Failed to load user function:', err);
});
