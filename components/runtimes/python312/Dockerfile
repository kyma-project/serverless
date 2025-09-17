FROM europe-docker.pkg.dev/kyma-project/prod/external/library/python:3.12.10-alpine3.21

# Serverless
LABEL source = git@github.com:kyma-project/serverless.git

# build-base and linux-headers are needed to install all requirements
RUN apk add --no-cache --virtual .build-deps build-base linux-headers

# cmake is required for some packages
RUN apk add --no-cache cmake

# TODO: copy to /usr/src/app
COPY ./requirements.txt /usr/src/app/function/requirements.txt
RUN chmod 644 /usr/src/app/function/requirements.txt

RUN pip install --no-cache-dir -r /usr/src/app/function/requirements.txt

COPY ./ /usr/src/app
RUN chmod -R 755 /lib
RUN chmod 644 /usr/src/app/server.py

WORKDIR /

USER 1000
# Tracing propagators are configured based on OTEL_PROPAGATORS env variable https://opentelemetry.io/docs/instrumentation/python/manual/#using-environment-variables
ENV OTEL_PROPAGATORS=tracecontext,baggage,b3multi
ENV OTEL_PYTHON_REQUESTS_EXCLUDED_URLS="healthz,favicon.ico,metrics"

CMD ["python", "/usr/src/app/server.py"]
