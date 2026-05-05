"""
Stub SDK for local development and IDE autocomplete.
At runtime, the real sdk module bundled with the python314 container is used instead.

Install: pip install kyma-sdk
"""
from flask import Request
from opentelemetry import trace


def get_request() -> Request:
    """Returns the current Flask request object."""
    raise RuntimeError("sdk stub: not available outside a running Kyma function")


def get_tracer() -> trace.Tracer:
    """Returns the configured OpenTelemetry tracer."""
    raise RuntimeError("sdk stub: not available outside a running Kyma function")


def emit_cloud_event(event_type: str, source: str, data, optional_attributes: dict = None):
    """Publishes a CloudEvent to the configured publisher proxy."""
    raise RuntimeError("sdk stub: not available outside a running Kyma function")
