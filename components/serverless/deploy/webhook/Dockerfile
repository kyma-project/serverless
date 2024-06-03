# image builder base on golang:1.22.3-alpine3.20
FROM golang@sha256:e62cde42bd32f7b0435c997d4048ee7b76042e31bde380898e1df97eabdb01fd as builder

ENV BASE_APP_DIR=/workspace/go/src/github.com/kyma-project/serverless \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR ${BASE_APP_DIR}

# Copy the go source
COPY . ${BASE_APP_DIR}/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o webhook-server ./components/serverless/cmd/webhook/main.go \
&& mkdir /app \
&& mv ./webhook-server /app/webhook-server

# get latest CA certs from alpine:3.17.4
FROM europe-docker.pkg.dev/kyma-project/prod/external/alpine:3.20.0 as certs
RUN apk add --no-cache ca-certificates

FROM scratch

LABEL source = git@github.com:kyma-project/serverless.git

COPY --from=builder /app /app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
USER 1000

ENTRYPOINT ["/app/webhook-server"]
