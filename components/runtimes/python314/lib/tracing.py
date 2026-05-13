from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace.export import SimpleSpanProcessor
from opentelemetry.sdk.trace.sampling import DEFAULT_ON
from opentelemetry.instrumentation.requests import RequestsInstrumentor


def setup(tracecollector_endpoint) -> trace.Tracer:
    """Sets up the OpenTelemetry tracer with the given trace collector endpoint.

    Tracing propagators are configured based on OTEL_PROPAGATORS env variable set in dockerfile.
    See: https://opentelemetry.io/docs/instrumentation/python/manual/#using-environment-variables
    """
    provider = TracerProvider(
        resource=Resource.create(),
        sampler=DEFAULT_ON,
    )

    if tracecollector_endpoint:
        span_processor = SimpleSpanProcessor(OTLPSpanExporter(endpoint=tracecollector_endpoint))
        provider.add_span_processor(span_processor)

    # Sets the global default tracer provider
    trace.set_tracer_provider(provider)

    # Auto instrument all outgoing requests via `requests` library
    RequestsInstrumentor().instrument()

    # Creates a tracer from the global tracer provider
    return trace.get_tracer("io.kyma-project.serverless")
