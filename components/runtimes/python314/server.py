import os
import sys
import pathlib

import flask
from gevent import pywsgi
import prometheus_client

# Make lib/ importable as a flat namespace so handlers can `from sdk import get_request`
sys.path.insert(0, str(pathlib.Path(__file__).parent / 'lib'))

from lib import tracing, module, sdk  # noqa: E402

func_name = os.getenv('FUNC_NAME', '')
func_namespace = os.getenv('FUNC_NAMESPACE', '')
func_runtime = os.getenv('FUNC_RUNTIME', 'python314')
server_host = os.getenv('SERVER_HOST', '0.0.0.0')
server_port = int(os.getenv('SERVER_PORT', '8080'))
server_numthreads = int(os.getenv('SERVER_NUMTHREADS', '50'))
server_call_timeout = float(os.getenv('SERVER_CALL_TIMEOUT', '180'))
handler_folder = os.getenv('HANDLER_FOLDER', '/')
handler_module_name = os.getenv('HANDLER_MODULE_NAME', 'handler')
handler_function_name = os.getenv('HANDLER_FUNCTION_NAME', 'main')
trace_collector_endpoint = os.getenv('TRACE_COLLECTOR_ENDPOINT', '')
publisher_proxy_address = os.getenv('PUBLISHER_PROXY_ADDRESS', '')

print(f"Importing function sources from {handler_folder}/{handler_module_name}:{handler_function_name}")
print(f"Tracing configured with endpoint {trace_collector_endpoint}")
print(f"Publisher Proxy available on address {publisher_proxy_address}")
print(f"Starting {func_runtime} server {server_host}:{server_port}")
sys.stdout.flush()

tracer = tracing.setup(trace_collector_endpoint)
sdk._configure(tracer, publisher_proxy_address)

handler = module.Handler(handler_folder, handler_module_name, handler_function_name)

app = flask.Flask(__name__)


@app.route('/', methods=['GET', 'POST', 'PUT', 'HEAD', 'OPTIONS', 'DELETE'])
def userfunc_call():
    return handler.call(flask.request.method)


@app.errorhandler(500)
def internal_error(_error):
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


if __name__ == "__main__":
    pywsgi.WSGIServer(
        (server_host, server_port),
        app,
        spawn=server_numthreads,
        log=None,
    ).serve_forever()
