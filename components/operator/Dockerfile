#
# This Dockerfile is used to build serverless-operator image on every pre- and post-submit job
#


# Build the operator binary
FROM --platform=$BUILDPLATFORM europe-docker.pkg.dev/kyma-project/prod/external/library/golang:1.25.2-alpine3.22 AS builder

ARG TARGETOS TARGETARCH

WORKDIR /workdir

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY components/operator components/operator

# Build
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o operator ./components/operator/main.go


# Replace main images in values.yaml
# Bumping serverless images in the values.yaml - used for building "local" and "dev" images
FROM europe-docker.pkg.dev/kyma-project/prod/external/ghcr.io/nginx/alpine-fips:0.4.0-alpine3.22 AS replacer

WORKDIR /workdir

ARG PURPOSE
ARG IMG_DIRECTORY
ARG IMG_VERSION

ENV IMG_DIRECTORY=$IMG_DIRECTORY
ENV IMG_VERSION=$IMG_VERSION
ENV PROJECT_ROOT="."

RUN apk update && apk add bash yq

COPY components/operator/hack components/operator/hack
COPY config/serverless config/serverless
COPY config/buildless-serverless config/buildless-serverless
COPY hack/ hack/

RUN if [[ "dev" = "$PURPOSE" ]] ; then ./hack/replace_serverless_chart_images.sh all ; fi
RUN if [[ "local" = "$PURPOSE" ]] ; then ./hack/replace_serverless_chart_images.sh main-only ; fi
# do nothing (keep unchanged versions) for "release"


# Use distroless as minimal base image to package the operator binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --chown=65532:65532 --from=builder /workdir/operator .
COPY --chown=65532:65532 --from=replacer /workdir/config/serverless /module-chart
COPY --chown=65532:65532 --from=replacer /workdir/config/buildless-serverless /buildless-module-chart
USER 65532:65532

ENTRYPOINT ["/operator"]
