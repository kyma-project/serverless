'use strict';

const { AsyncLocalStorage } = require('async_hooks');
const { HTTP, CloudEvent } = require('cloudevents');
const axios = require('axios');

const _store = new AsyncLocalStorage();

let _tracer = null;
let _publisherProxyAddress = null;

function _configure(tracer, publisherProxyAddress) {
    _tracer = tracer;
    _publisherProxyAddress = publisherProxyAddress;
}

function runWithContext(req, res, fn) {
    return _store.run({ req, res }, fn);
}

function getRequest() {
    const ctx = _store.getStore();
    if (!ctx) throw new Error('getRequest() called outside of a request context');
    return ctx.req;
}

function getResponse() {
    const ctx = _store.getStore();
    if (!ctx) throw new Error('getResponse() called outside of a request context');
    return ctx.res;
}

function getTracer() {
    return _tracer;
}

function emitCloudEvent(type, source, data, optionalAttributes) {
    if (!_publisherProxyAddress) {
        throw new Error('sdk not configured: PUBLISHER_PROXY_ADDRESS is not set');
    }
    const attrs = Object.assign({ type, source, data }, optionalAttributes || {});
    if (!attrs.datacontenttype) {
        attrs.datacontenttype = typeof data === 'object' ? 'application/json' : 'text/plain';
    }
    const ce = new CloudEvent(attrs);
    const message = HTTP.structured(ce);
    return axios.post(_publisherProxyAddress, message.body, { headers: message.headers });
}

module.exports = { _configure, runWithContext, getRequest, getResponse, getTracer, emitCloudEvent };
