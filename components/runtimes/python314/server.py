import os

from lib import tracing, module

import flask
from gevent import pywsgi, Timeout
import prometheus_client
# TODO: tracing
# TODO: cloudevents
# TODO: monitoring
# TODO: sdk

# Configuration from environment variables
server_host=os.getenv('SERVER_HOST', '0.0.0.0')
server_port=int(os.getenv('SERVER_PORT', '8080'))
server_numthreads=int(os.getenv('SERVER_NUMTHREADS', '50'))
server_timeout=float(os.getenv('SERVER_TIMEOUT', '180'))
handler_module_folder=os.getenv('HANDLER_FOLDER', '/')
handler_module_name=os.getenv('HANDLER_MODULE_NAME', 'handler')
handler_module_function=os.getenv('HANDLER_FUNCTION_NAME', 'main')
tracecollector_endpoint = os.getenv('TRACE_COLLECTOR_ENDPOINT')
publisher_proxy_address = os.getenv('PUBLISHER_PROXY_ADDRESS')

app = flask.Flask(__name__)

handler = module.Handler(
    handler_module_folder, 
    handler_module_name, 
    handler_module_function,
)

# TODO: I've added PUT and OPTIONS methods. is it ok?
@app.route('/', methods=['GET', 'POST', 'PUT', 'HEAD', 'OPTIONS', 'DELETE'])
def userfunc_call():
    # TODO: implement user function call with timeout and error handling
    return handler.call(None)

@app.get('/favicon.ico')
def favicon():
    # TODO: serve a real favicon - for example redirect to kyma-project.io favicon
    return '', 204

@app.get('/healthz')
def healthz():
    return 'OK'

@app.get('/metrics')
def metrics():
    return prometheus_client.generate_latest(prometheus_client.REGISTRY), 200, {'Content-Type': prometheus_client.CONTENT_TYPE_LATEST}

# Run the Flask app using gevent WSGI server
if __name__ == "__main__":
    # TODO: check if we still need to setup loggedapp through WSGILogger
    # TODO: handle ctrl+c and SIGTERM signals to gracefully shutdown the server
    # TODO: implement request timeout handling - for example using gevent.Timeout
    # TODO: maybe we should run server using common target like `./run.sh`? or `make run-prod`? 
    #       to move whole code and deps related logic to one place and to build one common deploy for both runtimes
    pywsgi.WSGIServer(
        (server_host, server_port),
        app,
        spawn=server_numthreads,
        # TODO: do we need these logs?
        # log=None,
        # error_log=None,
    ).serve_forever()
