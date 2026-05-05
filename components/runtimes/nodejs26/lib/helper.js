'use strict';

const opentelemetry = require('@opentelemetry/api');

function configureGracefulShutdown(server) {
    let nextConnectionId = 0;
    const connections = {};
    let terminating = false;

    server.on('connection', connection => {
      const connectionId = nextConnectionId++;
      connection.$$isIdle = true;
      connections[connectionId] = connection;
      connection.on('close', () => delete connections[connectionId]);
    });

    server.on('request', (request, response) => {
      const connection = request.connection;
      connection.$$isIdle = false;

      response.on('finish', () => {
        connection.$$isIdle = true;
        if (terminating) {
          connection.destroy();
        }
      });
    });

    const handleShutdown = () => {
      console.log("Shutting down..");

      terminating = true;
      server.close(() => console.log("Server stopped"));

      for (const connectionId in connections) {
        if (connections.hasOwnProperty(connectionId)) {
          const connection = connections[connectionId];
          if (connection.$$isIdle) {
            connection.destroy();
          }
        }
      }
    };

    process.on('SIGINT', handleShutdown);
    process.on('SIGTERM', handleShutdown);
  }

function handleTimeOut(req, res, next){
  const timeout = Number(process.env.SERVER_CALL_TIMEOUT || '180');
  res.setTimeout(timeout * 1000, function(){
    res.sendStatus(408);
  });
  next();
}

const isFunction = (func) => {
  return func && func.constructor && func.call && func.apply;
};

const isPromise = (promise) => {
  return typeof promise.then == "function"
}


function handleError(err, span, sendResponse) {
    console.error(err);
    const errTxt = resolveErrorMsg(err);
    if (span) {
        span.setStatus({ code: opentelemetry.SpanStatusCode.ERROR, message: errTxt });
        span.setAttribute("error", errTxt);
    }
    sendResponse(errTxt, 500);
}

function resolveErrorMsg(_err) {
    return "Internal server error";
}

module.exports = {
  configureGracefulShutdown,
  handleTimeOut,
  isFunction,
  isPromise, 
  handleError
};