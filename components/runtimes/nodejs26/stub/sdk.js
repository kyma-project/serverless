/**
 * Stub SDK for local development and IDE autocomplete.
 * At runtime, the real sdk module bundled with the nodejs26 container is used instead.
 *
 * Install: npm install @kyma-project/sdk
 */

/** @returns {import('express').Request} */
function getRequest() {
    throw new Error('sdk stub: not available outside a running Kyma function');
}

/** @returns {import('express').Response} */
function getResponse() {
    throw new Error('sdk stub: not available outside a running Kyma function');
}

/** @returns {import('@opentelemetry/api').Tracer} */
function getTracer() {
    throw new Error('sdk stub: not available outside a running Kyma function');
}

/**
 * @param {string} type
 * @param {string} source
 * @param {*} data
 * @param {object} [optionalAttributes]
 */
function emitCloudEvent(type, source, data, optionalAttributes) {
    throw new Error('sdk stub: not available outside a running Kyma function');
}

module.exports = { getRequest, getResponse, getTracer, emitCloudEvent };
