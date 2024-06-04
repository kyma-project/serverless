FROM python:3.12.3-slim-bookworm

# Serverless
LABEL source = git@github.com:kyma-project/serverless.git

# build-base and linux-headers are needed to install all requirements
RUN apt update && apt install -y build-essential linux-headers-generic && \
    rm -rf /var/lib/apt/lists/*

COPY ./lib/requirements.txt /kubeless/requirements.txt
RUN chmod 644 /kubeless/requirements.txt

RUN pip install --no-cache-dir -r /kubeless/requirements.txt

COPY ./lib /
RUN chmod -R 755 /lib
COPY ./kubeless.py /
RUN chmod 644 /kubeless.py

WORKDIR /

USER 1000
# Tracing propagators are configured based on OTEL_PROPAGATORS env variable https://opentelemetry.io/docs/instrumentation/python/manual/#using-environment-variables
ENV OTEL_PROPAGATORS=tracecontext,baggage,b3multi
ENV OTEL_PYTHON_REQUESTS_EXCLUDED_URLS="healthz,favicon.ico,metrics"

CMD ["python", "/kubeless.py"]
