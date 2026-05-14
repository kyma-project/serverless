const { HTTP, CloudEvent } = require('cloudevents');
const axios = require('axios');

let _tracer = null;
let _publisherProxyAddress = null;
let _funcName = '';
let _funcNamespace = '';
let _funcRuntime = '';
let _serverCallTimeout = 180;
let _reqMbLimit = 1;

function configure(tracer, publisherProxyAddress, funcName, funcNamespace, funcRuntime, serverCallTimeout, reqMbLimit) {
    _tracer = tracer;
    _publisherProxyAddress = publisherProxyAddress;
    _funcName = funcName;
    _funcNamespace = funcNamespace;
    _funcRuntime = funcRuntime;
    _serverCallTimeout = serverCallTimeout;
    _reqMbLimit = reqMbLimit;
}

function getTracer() {
    return _tracer;
}

function getCloudEvent(req) {
    const isCloudEventContentType = req.is('application/cloudevents+json');
    const hasCeHeaders = req.get('ce-type') && req.get('ce-source');
    if (!isCloudEventContentType && !hasCeHeaders) {
        return null;
    }
    try {
        return HTTP.toEvent({ headers: req.headers, body: req.body });
    } catch (e) {
        return null;
    }
}

function emitCloudEvent(type, source, data, optionalAttributes) {
    const attrs = Object.assign({ type, source }, optionalAttributes || {});
    if (!attrs.datacontenttype) {
        attrs.datacontenttype = typeof data === 'object' ? 'application/json' : 'text/plain';
    }
    const ce = new CloudEvent(Object.assign(attrs, { data }));
    const message = HTTP.structured(ce);
    return axios.post(_publisherProxyAddress, message.body, { headers: message.headers });
}

function getFunctionName()  { return _funcName; }
function getNamespace()     { return _funcNamespace; }
function getRuntime()       { return _funcRuntime; }
function getTimeout()       { return _serverCallTimeout; }
function getBodySizeLimit() { return _reqMbLimit; }

module.exports = {
    configure,
    getTracer,
    getCloudEvent,
    emitCloudEvent,
    getFunctionName,
    getNamespace,
    getRuntime,
    getTimeout,
    getBodySizeLimit,
};
