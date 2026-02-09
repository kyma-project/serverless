
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


# TODO: can we base event on flask.Request type and get rid of extensions and dict?
#       maybe we should redesign Event structure? 
class Event:
    """
    Event class to wrap the incomming request with cloudevent sdk, tracing, and pass it to user function

    :param req: Flask request object
    :type req: Request
    :param tracer: Tracer object for tracing
    :type tracer: opentelemetry.trace.Tracer
    :param publisher_proxy_address: Address of the publisher proxy to emit cloud events
    """
    def __init__(self, req:Request, tracer, publisher_proxy_address):
        self.ceHeaders = dict()
        self.tracer = tracer
        self.request = req
        self.publisher_proxy_address = publisher_proxy_address
        data = req.get_data()
        self.ceHeaders.update({
            'extensions': {'request': req}
        })

        if self.__is_cloud_event(req):
            ce_headers = self.__build_cloud_event_attributes(req, data)
            self.ceHeaders.update(ce_headers)
        else:
            if req.headers.get('content-type') == 'application/json':
                data = req.json
                self.ceHeaders.update({'data': data})
    
    def __getitem__(self, item):
        return self.ceHeaders[item]

    def __setitem__(self, name, value):
        self.ceHeaders[name] = value
    
    def emitCloudEvent(self, type, source, data, optionalCloudEventAttributes=None):
        """
        Emit a CloudEvent

        :param type: CloudEvent type
        :param source: CloudEvent source
        :param data: CloudEvent data
        :param optionalCloudEventAttributes: Optional CloudEvent attributes
        """
        attributes = {
            "type": type,
            "source": source,
        }
        if optionalCloudEventAttributes is not None:
            attributes.update(optionalCloudEventAttributes)

        event = CloudEvent(attributes, data)
        headers, body = to_structured(event)

        requests.post(self.publisher_proxy_address, data=body, headers=headers)

    
    def __build_cloud_event_attributes(self, req:Request):
        event = from_http(req.headers, req.get_data)
        ceHeaders = {
            'data': event.data,
            'ce-type': event['type'],
            'ce-source': event['source'],
            'ce-id': event['id'],
            'ce-time': event['time'],
        }
        if event.get('eventtypeversion') is not None:
            ceHeaders['ce-eventtypeversion'] = event.get('eventtypeversion')

        if event.get('specversion') is not None:
            ceHeaders['ce-specversion'] = event.get('specversion')
            
        return ceHeaders

    def __has_ce_headers(self, headers):
        has = 'ce-type' in headers and 'ce-source' in headers
        return has


    def __is_cloud_event(self, req:Request):
        return (req.content_type != None and 'application/cloudevents+json' in req.content_type.split(';')) or self.__has_ce_headers(req.headers)
