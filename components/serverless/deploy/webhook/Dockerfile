FROM --platform=$BUILDPLATFORM europe-docker.pkg.dev/kyma-project/prod/external/library/golang:1.25.5-alpine3.22 AS builder

ARG TARGETOS TARGETARCH

ENV BASE_APP_DIR=/workspace/go/src/github.com/kyma-project/serverless \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR ${BASE_APP_DIR}

# Copy the go source
COPY . ${BASE_APP_DIR}/

# Build
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOFIPS140=v1.0.0 go build -a -o webhook-server ./components/serverless/cmd/webhook/main.go \
    && mkdir /app \
    && mv ./webhook-server /app/webhook-server

# get latest CA certs from alpine:3.17.4
FROM europe-docker.pkg.dev/kyma-project/prod/external/library/alpine:3.23.0 AS certs
RUN apk add --no-cache ca-certificates

FROM scratch

LABEL source=git@github.com:kyma-project/serverless.git

ENV GODEBUG=fips140=only,tlsmlkem=0
COPY --from=builder /app /app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
USER 1000

ENTRYPOINT ["/app/webhook-server"]
