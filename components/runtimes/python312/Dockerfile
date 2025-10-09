FROM europe-docker.pkg.dev/kyma-project/prod/external/library/alpine:3.22.1

ARG OPENSSL_VERSION=3.5.3
ARG PYTHON_VERSION=3.12.11

# Serverless
LABEL source = git@github.com:kyma-project/serverless.git

# Install OpenSSL FIPS
RUN set -ex \
    && cpus=$(nproc --all) \
    && apk add --no-cache --virtual .build-deps make gcc libgcc musl-dev linux-headers perl vim \
    && wget https://www.openssl.org/source/openssl-${OPENSSL_VERSION}.tar.gz \
    && tar -xf openssl-${OPENSSL_VERSION}.tar.gz \
    && cd openssl-${OPENSSL_VERSION} \
    && ./Configure enable-fips shared enable-ec_nistp_64_gcc_128 \
    && make -j${cpus} \
    && make -j${cpus} install \
    && make -j${cpus} install_fips \
    && cp -a /etc/ssl/certs/* /usr/local/ssl/certs/ \
    && cp -a /etc/ssl/cert.pem /usr/local/ssl/ \
    && apk del .build-deps \
    && cd .. \
    && rm -rf openssl-${OPENSSL_VERSION}.tar.gz openssl-${OPENSSL_VERSION}

COPY openssl.cnf /etc/ssl/openssl.cnf

ENV OPENSSL_FIPS=1
ENV OPENSSL_DIR=/usr/local
ENV OPENSSL_CONF=/etc/ssl/openssl.cnf
ENV OPENSSL_MODULES=/usr/local/lib64/ossl-modules

ENV LDFLAGS="-L/usr/local/lib/ -L/usr/local/lib64/"
ENV LD_LIBRARY_PATH="/usr/local/lib/:/usr/local/lib64/"
ENV CPPFLAGS="-I/usr/local/include -I/usr/local/include/openssl"
ENV CFLAGS="-I/usr/local/include"

# Install Python
RUN set -ex \
    && apk add --no-cache --virtual .build-deps make gcc libgcc musl-dev linux-headers perl vim tcl-dev tk-dev bzip2-dev zlib-dev libffi-dev cups-dev krb5-dev yaml-dev sqlite-dev libffi-dev libxml2-dev libxslt-dev \
    && wget https://www.python.org/ftp/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tgz \
    && tar -xvf Python-${PYTHON_VERSION}.tgz \
    && cd Python-${PYTHON_VERSION} \
    && ./configure --enable-optimizations --enable-shared --with-ensurepip=no --without-builtin-hashlib-hashes \
    && make \
    && make install \
    && cd .. \
    && apk del .build-deps \
    && ln -sf /usr/local/bin/idle3 /usr/local/bin/idle \
    && ln -sf /usr/local/bin/pydoc3 /usr/local/bin/pydoc \
    && ln -sf /usr/local/bin/python3 /usr/local/bin/python \
    && ln -sf /usr/local/bin/python3-config /usr/local/bin/python-config \
    && rm -f Python-${PYTHON_VERSION}.tgz \
    && rm -rf Python-${PYTHON_VERSION}

# Install deps
RUN set -ex \
    && apk add --no-cache libintl libffi rust cargo pkgconfig

RUN set -ex \
    && apk add --no-cache curl \
    && curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py \
    && python3 get-pip.py \
    && rm get-pip.py \
    && pip install cryptography --no-binary cryptography

# build-base and linux-headers are needed to install all requirements
RUN apk add --no-cache --virtual .build-deps build-base linux-headers

# cmake is required for some packages
RUN apk add --no-cache cmake

COPY ./requirements.txt /kubeless/requirements.txt
RUN chmod 644 /kubeless/requirements.txt

RUN pip install --no-cache-dir -r /kubeless/requirements.txt

COPY ./ /
RUN chmod -R 755 /lib
RUN chmod 644 /server.py

WORKDIR /

USER 1000
# Tracing propagators are configured based on OTEL_PROPAGATORS env variable https://opentelemetry.io/docs/instrumentation/python/manual/#using-environment-variables
ENV OTEL_PROPAGATORS=tracecontext,baggage,b3multi
ENV OTEL_PYTHON_REQUESTS_EXCLUDED_URLS="healthz,favicon.ico,metrics"

CMD ["python", "/server.py"]
