import requests as _requests
from flask import request as _flask_request
from cloudevents.core.v1.event import CloudEvent
from cloudevents.core.bindings.http import to_structured_event
from opentelemetry import trace as _trace

_tracer = None
_publisher_proxy_address = None


def _configure(tracer, publisher_proxy_address):
    global _tracer, _publisher_proxy_address
    _tracer = tracer
    _publisher_proxy_address = publisher_proxy_address


def get_request():
    return _flask_request


def get_tracer() -> _trace.Tracer:
    return _tracer


def emit_cloud_event(event_type: str, source: str, data, optional_attributes=None):
    if _publisher_proxy_address is None:
        raise RuntimeError("sdk not configured: PUBLISHER_PROXY_ADDRESS is not set")
    attributes = {"type": event_type, "source": source}
    if optional_attributes is not None:
        attributes.update(optional_attributes)
    event = CloudEvent(attributes, data)
    message = to_structured_event(event)
    response = _requests.post(_publisher_proxy_address, data=message.body, headers=message.headers)
    response.raise_for_status()
    return response
