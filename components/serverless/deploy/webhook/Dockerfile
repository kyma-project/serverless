# image builder base on golang:1.21.4-alpine3.18
FROM golang@sha256:f475434ea2047a83e9ba02a1da8efc250fa6b2ed0e9e8e4eb8c5322ea6997795 as builder

ENV BASE_APP_DIR=/workspace/go/src/github.com/kyma-project/kyma/components/function-controller \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR ${BASE_APP_DIR}

# Copy the go source
COPY . ${BASE_APP_DIR}/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o webhook-server cmd/webhook/main.go \
&& mkdir /app \
&& mv ./webhook-server /app/webhook-server

# get latest CA certs from alpine:3.17.4
FROM europe-docker.pkg.dev/kyma-project/prod/external/alpine@sha256:e95676db9e4a4f16f6cc01a8915368f82b018cc07aba951c1bd1db586c081388 as certs
RUN apk add --no-cache ca-certificates

FROM scratch

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /app /app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
USER 1000

ENTRYPOINT ["/app/webhook-server"]
