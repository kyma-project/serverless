FROM python:3.14-alpine

# Serverless
LABEL source = git@github.com:kyma-project/serverless.git

# build-base and linux-headers are needed to install all requirements
RUN apk add --no-cache --virtual .build-deps build-base linux-headers

RUN mkdir -p /usr/src/app
RUN mkdir -p /usr/src/app/lib
WORKDIR /usr/src/app

COPY ./requirements.txt ./requirements.txt
RUN chmod 644 ./requirements.txt

RUN pip install --no-cache-dir -r ./requirements.txt

COPY ./ ./
RUN chmod -R 755 ./lib
RUN chmod 644 ./server.py

USER 1000

# Tracing propagators are configured based on OTEL_PROPAGATORS env variable https://opentelemetry.io/docs/instrumentation/python/manual/#using-environment-variables
ENV OTEL_PROPAGATORS=tracecontext,baggage,b3multi
ENV OTEL_PYTHON_REQUESTS_EXCLUDED_URLS="healthz,favicon.ico,metrics"

# Set PYTHONUNBUFFERED to ensure that logs are flushed immediately
ENV PYTHONUNBUFFERED=TRUE

CMD ["python", "./server.py"]
