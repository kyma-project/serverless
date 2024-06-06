FROM golang:1.22.3-alpine3.20 as builder

ENV BASE_APP_DIR=/workspace/go/src/github.com/kyma-project/serverless \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR ${BASE_APP_DIR}

#
# copy files allowed in .dockerignore
#
COPY . ${BASE_APP_DIR}/

RUN go build -ldflags "-s -w" -a -o manager ./components/serverless/cmd/manager/main.go \
    && mkdir /app \
    && mv ./manager /app/manager

FROM europe-docker.pkg.dev/kyma-project/prod/external/alpine:3.20.0 as certs

RUN apk add --no-cache ca-certificates

FROM scratch

LABEL source = git@github.com:kyma-project/serverless.git

COPY --from=builder /app /app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/app/manager"]

