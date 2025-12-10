FROM cgr.dev/sap.com/python-fips:3.13-dev

# Serverless
LABEL source = git@github.com:kyma-project/serverless.git

USER root

# build-base and linux-headers are needed to install all requirements
RUN apk add --no-cache --virtual .build-deps build-base linux-headers

# cmake is required for some packages
RUN apk add --no-cache cmake

COPY ./requirements.txt /usr/src/app/kubeless/requirements.txt
RUN chmod 644 /usr/src/app/kubeless/requirements.txt

RUN pip install --no-cache-dir -r /usr/src/app/kubeless/requirements.txt

COPY ./ /usr/src/app
RUN chmod -R 755 /usr/src/app/lib
RUN chmod 644 /usr/src/app/server.py

WORKDIR /usr/src/app

USER 1000
# Tracing propagators are configured based on OTEL_PROPAGATORS env variable https://opentelemetry.io/docs/instrumentation/python/manual/#using-environment-variables
ENV OTEL_PROPAGATORS=tracecontext,baggage,b3multi
ENV OTEL_PYTHON_REQUESTS_EXCLUDED_URLS="healthz,favicon.ico,metrics"

CMD ["python", "/usr/src/app/server.py"]