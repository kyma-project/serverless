import ce from './lib/ce.js';
import helper from './lib/helper.js';
import bodyParser from 'body-parser';
import process from 'process';
import morgan from "morgan";

import { setupTracer, getCurrentSpan } from './lib/tracer.js';
import { getMetrics, setupMetrics, createFunctionDurationHistogram, createFunctionCallsTotalCounter, createFunctionFailuresTotalCounter  } from './lib/metrics.js';


// To catch unhandled exceptions thrown by user code async callbacks,
// these exceptions cannot be catched by try-catch in user function invocation code below
process.on("uncaughtException", (err) => {
    console.error(`Caught exception: ${err}`);
});

const serviceNamespace = process.env.SERVICE_NAMESPACE;
const functionName = process.env.FUNC_NAME;
const bodySizeLimit = Number(process.env.REQ_MB_LIMIT || '1');
const funcPort = Number(process.env.FUNC_PORT || '8080');
const timeout = Number(process.env.FUNC_TIMEOUT || '180'); // Default to 180 seconds

const tracer = setupTracer(functionName);
setupMetrics(functionName);

const callsTotalCounter = createFunctionCallsTotalCounter(functionName);
const failuresTotalCounter = createFunctionFailuresTotalCounter(functionName);
const durationHistogram = createFunctionDurationHistogram(functionName);

//require express must be called AFTER tracer was setup!!!!!!
import express from "express";
const app = express();


// User function.  Starts out undefined.
let userFunction;


// Request logger
if (process.env["KYMA_INTERNAL_LOGGER_ENABLED"]) {
    app.use(morgan("combined"));
}


app.use(bodyParser.json({ type: ['application/json', 'application/cloudevents+json'], limit: `${bodySizeLimit}mb`, strict: false  }))
app.use(bodyParser.text({ type: ['text/*'], limit: `${bodySizeLimit}mb`  }))
app.use(bodyParser.urlencoded({ limit: `${bodySizeLimit}mb`, extended: true }));
app.use(bodyParser.raw({limit: `${bodySizeLimit}mb`, type: () => true}))

app.use(helper.handleTimeOut);

app.get("/healthz", (req, res) => {
    res.status(200).send("OK")
})

app.get("/metrics", (req, res) => {
    getMetrics(req, res)
})

app.get('/favicon.ico', (req, res) => res.status(204));

// Generic route -- all http requests go to the user function.
// Since express 5.0.0 '*' wildcard requires a named variable (here: `path`)
// see https://github.com/pillarjs/path-to-regexp?tab=readme-ov-file#errors
app.all("*path", (req, res, next) => {


    res.header('Access-Control-Allow-Origin', '*');
    if (req.method === 'OPTIONS') {
        // CORS preflight support (Allow any method or header requested)
        res.header('Access-Control-Allow-Methods', req.headers['access-control-request-method']);
        res.header('Access-Control-Allow-Headers', req.headers['access-control-request-headers']);
        res.end();
    } else {
    
        callsTotalCounter.add(1)
        const startTime = new Date().getTime()

        if (!userFunction) {
            failuresTotalCounter.add(1)
            res.status(500).send("User function not loaded");
            return;
        }

        const event = ce.buildEvent(req, res, tracer);

        const context = {
            'function-name': functionName,
            'runtime': process.env.FUNC_RUNTIME,
            'namespace': serviceNamespace,
            'timeout': timeout,
            'body-size-limit': bodySizeLimit
        };

        const sendResponse = (body, status, headers) => {
            if (res.writableEnded) return;
            if (headers) {
                for (let name of Object.keys(headers)) {
                    res.set(name, headers[name]);
                }
            }
            if(body){
                if(status){
                    res.status(status);
                } 
                switch (typeof body) {
                    case 'object':
                        res.json(body); // includes res.end(), null also handled
                        break;
                    case 'undefined':
                        res.end();
                        break;
                    default:
                        res.end(body);
                }
            } else if(status){
                res.sendStatus(status);
            } else {
                res.end();
            }
        };

        const currentSpan = getCurrentSpan();

        try {
            // Execute the user function
            const out = userFunction(event, context, sendResponse);
            if (out && helper.isPromise(out)) {
                out.then(result => {
                    sendResponse(result)
                })
                .catch((err) => {
                    helper.handleError(err, currentSpan, sendResponse)
                    failuresTotalCounter.add(1);
                })
            } else {
                sendResponse(out);
            }
        } catch (err) {
            helper.handleError(err, currentSpan, sendResponse)
            failuresTotalCounter.add(1);
        }

        const endTime = new Date().getTime()
        const executionTime = endTime - startTime;
        durationHistogram.record(executionTime);
    }
});


const server = app.listen(funcPort);

helper.configureGracefulShutdown(server);

let startTime = process.hrtime();
import('./function/handler.js').then((fn) => {
    if (helper.isFunction(fn.main)) {
        userFunction = fn.main
        let elapsed = process.hrtime(startTime);
        console.log(
            `user code loaded in ${elapsed[0]}sec ${elapsed[1] / 1000000}ms`
        );
    } else {
        console.error("Content loaded is not a function. Make sure your function exports 'main' function", fn)
    }
});
