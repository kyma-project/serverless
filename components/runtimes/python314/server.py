import os

import flask
from gevent import pywsgi
import prometheus_client

from lib import tracing, module, sdk

func_name = os.getenv('FUNC_NAME', '')
func_namespace = os.getenv('FUNC_NAMESPACE', '')
func_runtime = os.getenv('FUNC_RUNTIME', 'python314')
server_host = os.getenv('SERVER_HOST', '0.0.0.0')
server_port = int(os.getenv('SERVER_PORT', '8080'))
server_numthreads = int(os.getenv('SERVER_NUMTHREADS', '50'))
server_call_timeout = float(os.getenv('SERVER_CALL_TIMEOUT', '180'))
handler_folder = os.getenv('HANDLER_FOLDER', '/kubeless')
handler_module_name = os.getenv('HANDLER_MODULE_NAME', 'handler')
handler_function_name = os.getenv('HANDLER_FUNCTION_NAME', 'main')
trace_collector_endpoint = os.getenv('TRACE_COLLECTOR_ENDPOINT', '')
publisher_proxy_address = os.getenv('PUBLISHER_PROXY_ADDRESS', '')

print(f"Importing function sources from {handler_folder}/{handler_module_name}:{handler_function_name}", flush=True)
print(f"Tracing configured with endpoint {trace_collector_endpoint}", flush=True)
print(f"Publisher Proxy available on address {publisher_proxy_address}", flush=True)
print(f"Starting {func_runtime} server {server_host}:{server_port}", flush=True)

tracer = tracing.setup(trace_collector_endpoint)
sdk._configure(tracer, publisher_proxy_address, func_name, func_namespace, func_runtime, server_call_timeout)

handler = module.Handler(handler_folder, handler_module_name, handler_function_name)

app = flask.Flask(__name__)


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


if __name__ == '__main__':
    pywsgi.WSGIServer(
        (server_host, server_port),
        app,
        spawn=server_numthreads,
        log=None,
    ).serve_forever()
