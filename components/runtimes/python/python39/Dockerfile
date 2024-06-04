FROM python:3.9.19-alpine3.20

# Serverless
LABEL source = git@github.com:kyma-project/serverless.git

# build-base and linux-headers are needed to install all requirements
RUN apk add --no-cache --virtual .build-deps build-base linux-headers

COPY ./python39/requirements.txt /kubeless/requirements.txt
RUN chmod 644 /kubeless/requirements.txt

RUN pip install --no-cache-dir -r /kubeless/requirements.txt

COPY ./ /
RUN chmod -R 755 /lib
RUN chmod 644 /kubeless.py

WORKDIR /

USER 1000
# Tracing propagators are configured based on OTEL_PROPAGATORS env variable https://opentelemetry.io/docs/instrumentation/python/manual/#using-environment-variables
ENV OTEL_PROPAGATORS=tracecontext,baggage,b3multi
ENV OTEL_PYTHON_REQUESTS_EXCLUDED_URLS="healthz,favicon.ico,metrics"

CMD ["python", "/kubeless.py"]
