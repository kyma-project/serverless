import requests as _requests
from flask import request as _flask_request
from cloudevents.core.v1.event import CloudEvent
from cloudevents.core.bindings.http import from_http_event, to_structured_event, HTTPMessage

_tracer = None
_publisher_proxy_address = None
_func_name = ''
_func_namespace = ''
_func_runtime = ''
_server_call_timeout = 180
_func_body_mb_limit = 100


def _configure(tracer, publisher_proxy_address, func_name, func_namespace, func_runtime, server_call_timeout, func_body_mb_limit):
    global _tracer, _publisher_proxy_address, _func_name, _func_namespace, _func_runtime, _server_call_timeout, _func_body_mb_limit
    _tracer = tracer
    _publisher_proxy_address = publisher_proxy_address
    _func_name = func_name
    _func_namespace = func_namespace
    _func_runtime = func_runtime
    _server_call_timeout = server_call_timeout
    _func_body_mb_limit = func_body_mb_limit


def get_tracer():
    return _tracer


def get_cloud_event():
    req = _flask_request
    content_type = req.content_type or ''
    has_ce_content_type = 'application/cloudevents+json' in content_type.split(';')
    has_ce_headers = 'ce-type' in req.headers and 'ce-source' in req.headers
    if not (has_ce_content_type or has_ce_headers):
        return None
    message = HTTPMessage(headers=dict(req.headers), body=req.get_data())
    return from_http_event(message)


def emit_cloud_event(type, source, data, optional_attributes=None):
    attributes = {'type': type, 'source': source}
    if optional_attributes:
        attributes.update(optional_attributes)
    event = CloudEvent(attributes, data)
    message = to_structured_event(event)
    _requests.post(_publisher_proxy_address, data=message.body, headers=message.headers)


def get_function_name():
    return _func_name


def get_namespace():
    return _func_namespace


def get_runtime():
    return _func_runtime


def get_timeout():
    return _server_call_timeout


def get_body_size_limit():
    return _func_body_mb_limit
