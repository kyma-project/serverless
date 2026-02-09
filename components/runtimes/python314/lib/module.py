
import sys
import importlib

import requests
import prometheus_client as prom
from cloudevents.http import from_http, CloudEvent
from cloudevents.conversion import to_structured
from flask import Request

class Handler:
    """
    Handler class to import user function from module and call it with context
    
    It also:
    * wraps the incomming request with cloudevent sdk, tracing, and passes it to user function
    * registers prometheus metrics for user function calls, duration and errors

    :param module_folder: folder where user module is located
    :param module_name: name of user module without .py extension
    :param module_function_name: name of user function in the module
    """
    def __init__(self, module_folder, module_name, module_function_name):
        # import user function from module and store it
        sys.path.append(module_folder)
        module = importlib.import_module(module_name)
        self.func = getattr(module, module_function_name)
        
        # register prometheus metrics for user function
        self.func_hist = prom.Histogram(
            'function_duration_seconds', 'Duration of user function in seconds', ['method']
        )
        self.func_calls = prom.Counter(
            'function_calls_total', 'Number of calls to user function', ['method']
        )
        self.func_errors = prom.Counter(
            'function_failures_total', 'Number of exceptions in user function', ['method']
        )

    def call(self, event):
        """
        Call the user function with context and measure metrics
        
        :param event: Event object passed to user function
        """
        method = event.request.method
        self.func_calls.labels(method).inc()
        with self.func_errors.labels(method).count_exceptions():
            with self.func_hist.labels(method).time():
                self.func_calls.labels(method).inc()
                # TODO: do we need context?
                return self.func(event, None)    
