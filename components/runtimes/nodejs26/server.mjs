import sdk from './lib/sdk.js';
import helper from './lib/helper.js';
import bodyParser from 'body-parser';
import process from 'process';

import { setupTracer, getCurrentSpan } from './lib/tracer.js';
import { getMetrics, setupMetrics, createFunctionDurationHistogram, createFunctionCallsTotalCounter, createFunctionFailuresTotalCounter } from './lib/metrics.js';

process.on("uncaughtException", (err) => {
    console.error(`Caught exception: ${err}`);
});

const funcName = process.env.FUNC_NAME || '';
const funcNamespace = process.env.FUNC_NAMESPACE || '';
const funcRuntime = process.env.FUNC_RUNTIME || 'nodejs26';
const serverHost = process.env.SERVER_HOST || '0.0.0.0';
const serverPort = parseInt(process.env.SERVER_PORT || '8080', 10);
const serverCallTimeout = Number(process.env.SERVER_CALL_TIMEOUT || '180');
const handlerPath = process.env.HANDLER_PATH || './handler.js';
const traceCollectorEndpoint = process.env.TRACE_COLLECTOR_ENDPOINT || '';
const publisherProxyAddress = process.env.PUBLISHER_PROXY_ADDRESS || '';

console.log(`Importing function sources from ${handlerPath}:main`);
console.log(`Tracing configured with endpoint ${traceCollectorEndpoint}`);
console.log(`Publisher Proxy available on address ${publisherProxyAddress}`);
console.log(`Starting ${funcRuntime} server ${serverHost}:${serverPort}`);

const tracer = setupTracer(funcName);
setupMetrics(funcName);
sdk._configure(tracer, publisherProxyAddress);

const callsTotalCounter = createFunctionCallsTotalCounter(funcName);
const failuresTotalCounter = createFunctionFailuresTotalCounter(funcName);
const durationHistogram = createFunctionDurationHistogram(funcName);

// require express AFTER tracer setup
import express from "express";
const app = express();

let userFunction;

app.use(bodyParser.json({ type: ['application/json', 'application/cloudevents+json'], limit: '1mb', strict: false }));
app.use(bodyParser.text({ type: ['text/*'], limit: '1mb' }));
app.use(bodyParser.urlencoded({ limit: '1mb', extended: true }));
app.use(bodyParser.raw({ limit: '1mb', type: () => true }));

app.use(helper.handleTimeOut);

app.get("/healthz", (req, res) => res.status(200).send("OK"));
app.get("/metrics", (req, res) => getMetrics(req, res));
app.get('/favicon.ico', (req, res) => res.status(204).end());

app.all("*path", (req, res) => {
    res.header('Access-Control-Allow-Origin', '*');

    if (req.method === 'OPTIONS') {
        res.header('Access-Control-Allow-Methods', req.headers['access-control-request-method']);
        res.header('Access-Control-Allow-Headers', req.headers['access-control-request-headers']);
        res.end();
        return;
    }

    callsTotalCounter.add(1);
    const startTime = new Date().getTime();

    if (!userFunction) {
        failuresTotalCounter.add(1);
        res.status(500).send("User function not loaded");
        return;
    }

    const currentSpan = getCurrentSpan();

    sdk.runWithContext(req, res, () => {
        try {
            const out = userFunction();
            if (out && helper.isPromise(out)) {
                out.then(result => {
                    if (!res.writableEnded) res.json(result);
                }).catch(err => {
                    helper.handleError(err, currentSpan, (body, status) => {
                        if (!res.writableEnded) res.status(status || 500).send(body);
                    });
                    failuresTotalCounter.add(1);
                });
            } else if (out !== undefined && !res.writableEnded) {
                res.json(out);
            }
        } catch (err) {
            helper.handleError(err, currentSpan, (body, status) => {
                if (!res.writableEnded) res.status(status || 500).send(body);
            });
            failuresTotalCounter.add(1);
        }
    });

    const endTime = new Date().getTime();
    durationHistogram.record(endTime - startTime);
});

const server = app.listen(serverPort, serverHost);
helper.configureGracefulShutdown(server);

let startTime = process.hrtime();
import(handlerPath).then((fn) => {
    if (helper.isFunction(fn.main)) {
        userFunction = fn.main;
        const elapsed = process.hrtime(startTime);
        console.log(`user code loaded in ${elapsed[0]}sec ${elapsed[1] / 1000000}ms`);
    } else {
        console.error("Content loaded is not a function. Make sure your function exports 'main' function", fn);
    }
}).catch((err) => {
    console.error("Failed to load user function:", err);
});
