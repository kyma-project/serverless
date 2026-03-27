FROM alpine:3.21

RUN apk add --no-cache \
    bash \
    curl \
    gettext \
    && curl -LO "https://dl.k8s.io/release/$(curl -Ls https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" \
    && install -m 0755 kubectl /usr/local/bin/kubectl \
    && rm kubectl

WORKDIR /test

COPY . .

ENTRYPOINT ["./test_runtimes.sh"]
