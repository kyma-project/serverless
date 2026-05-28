import sys
import importlib

import gevent
import prometheus_client as prom


class Handler:
    """Imports user function from module and calls it with timeout enforcement.

    Registers prometheus metrics for user function calls, duration and errors.
    """
    def __init__(self, module_folder, module_name, module_function_name, timeout):
        # import user function from module and store it
        sys.path.append(module_folder)
        module = importlib.import_module(module_name)
        self.func = getattr(module, module_function_name)
        self.timeout = timeout

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

    def call(self, method):
        self.func_calls.labels(method).inc()
        with self.func_errors.labels(method).count_exceptions():
            with self.func_hist.labels(method).time():
                try:
                    with gevent.Timeout(self.timeout):
                        return self.func()
                except gevent.Timeout:
                    return 'Timeout while processing the function', 408
