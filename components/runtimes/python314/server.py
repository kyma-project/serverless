import logging
import os
import sys

# Add lib/ to sys.path before any imports so user handlers can `import sdk` using
# the same module instance that server.py configures (lib.sdk vs sdk would be separate).
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), 'lib'))

import flask
from gevent import pywsgi
import prometheus_client
import requestlogger

import sdk
from lib import tracing, module

# Configuration from environment variables
func_namespace = os.getenv('SERVICE_NAMESPACE', '')
func_name = os.getenv('FUNC_NAME', '')
func_runtime = os.getenv('FUNC_RUNTIME', 'python314')
server_host = os.getenv('SERVER_HOST', '0.0.0.0')
server_port = int(os.getenv('SERVER_PORT', '8080'))
server_numthreads = int(os.getenv('SERVER_NUMTHREADS', '50'))
func_timeout = int(os.getenv('FUNC_TIMEOUT', '180'))
func_body_mb_limit = int(os.getenv('FUNC_BODY_MB_LIMIT', '100'))
handler_folder = os.getenv('HANDLER_PATH', '/')
handler_module_name = os.getenv('HANDLER_MOD_NAME', 'handler')
handler_function_name = os.getenv('HANDLER_FUNC_NAME', 'main')
trace_collector_endpoint = os.getenv('TRACE_COLLECTOR_ENDPOINT', '')
publisher_proxy_address = os.getenv('PUBLISHER_PROXY_ADDRESS', '')

# TODO(Hx2): check if this is really necessary for the kyma cli eject
function_location = os.getenv('FUNCTION_PATH', default='/')
sys.path.append(function_location)

print(f"Importing function sources from {handler_folder}/{handler_module_name}:{handler_function_name}", flush=True)
print(f"Tracing configured with endpoint {trace_collector_endpoint}", flush=True)
print(f"Publisher Proxy available on address {publisher_proxy_address}", flush=True)
print(f"Starting {func_runtime} server {server_host}:{server_port}", flush=True)

tracer = tracing.setup(trace_collector_endpoint)
sdk._configure(tracer, publisher_proxy_address, func_name, func_namespace, func_runtime, func_timeout, func_body_mb_limit)

handler = module.Handler(handler_folder, handler_module_name, handler_function_name, func_timeout)

app = flask.Flask(__name__)
app.config['MAX_CONTENT_LENGTH'] = func_body_mb_limit * 1024 * 1024

if os.getenv('SERVER_INTERNAL_LOGGER'):
    wsgi_app = requestlogger.WSGILogger(
        app,
        [logging.StreamHandler(stream=sys.stdout)],
        requestlogger.ApacheFormatter(),
    )
else:
    wsgi_app = app


@app.route('/', defaults={'path': ''}, methods=['GET', 'POST', 'PUT', 'HEAD', 'OPTIONS', 'DELETE', 'PATCH'])
@app.route('/<path:path>', methods=['GET', 'POST', 'PUT', 'HEAD', 'OPTIONS', 'DELETE', 'PATCH'])
def userfunc_call(path=''):
    return handler.call(flask.request.method)


@app.errorhandler(500)
def internal_error(error):
    return 'Internal Server Error', 500


@app.get('/favicon.ico')
def favicon():
    return '', 204


@app.get('/healthz')
def healthz():
    return 'OK', 200


@app.get('/metrics')
def metrics():
    return prometheus_client.generate_latest(prometheus_client.REGISTRY), 200, {'Content-Type': prometheus_client.CONTENT_TYPE_LATEST}

# Run the Flask app using gevent WSGI server
if __name__ == '__main__':
    import gevent
    import signal

    server = pywsgi.WSGIServer(
        (server_host, server_port),
        wsgi_app,
        spawn=server_numthreads,
        log=None,
    )

    def shutdown():
        print('Shutting down..', flush=True)
        server.stop()

    gevent.signal_handler(signal.SIGTERM, shutdown)
    gevent.signal_handler(signal.SIGINT, shutdown)

    server.serve_forever()
