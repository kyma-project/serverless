#
# This Dockerfile is used to build dockerregistry-operator image on every pre- and post-submit job
#


# Build the operator binary
FROM golang:1.22.1 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY components/operator components/operator

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o operator ./components/operator/main.go


# Replace main images in values.yaml
# Bumping dockerregistry images in the values.yaml - used for building "local" and "dev" images
FROM alpine:3.19.1 as replacer

WORKDIR /workspace

# TODO: remove variable PURPOSE
ARG PURPOSE
ARG IMG_DIRECTORY
ARG IMG_VERSION

ENV IMG_DIRECTORY=$IMG_DIRECTORY
ENV IMG_VERSION=$IMG_VERSION
ENV PROJECT_ROOT="."

RUN apk update && apk add bash yq

COPY components/operator/hack components/operator/hack
COPY config/docker-registry config/docker-registry
COPY hack/ hack/


# Use distroless as minimal base image to package the operator binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --chown=65532:65532 --from=builder /workspace/operator .
COPY --chown=65532:65532 --from=replacer /workspace/config/docker-registry /module-chart
USER 65532:65532

ENTRYPOINT ["/operator"]
