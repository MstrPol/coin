# syntax=docker/dockerfile:1.8
ARG REGISTRY=""

FROM ${REGISTRY}golang:1.25-alpine AS build
WORKDIR /src/coin-api

ARG GOPROXY="https://proxy.golang.org,direct"
ARG GOPRIVATE=""
ARG GOINSECURE="*"
ARG GIT_SSL_NO_VERIFY=true

ENV GOINSECURE=${GOINSECURE} \
    GIT_SSL_NO_VERIFY=${GIT_SSL_NO_VERIFY}

COPY coin-api/go.mod coin-api/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=secret,id=go_auth \
    if [ -f /run/secrets/go_auth ]; then cat /run/secrets/go_auth > ~/.netrc; fi && \
    go mod download

COPY coin-api/ .

ARG VERSION=0.1.0-dev
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    VERSION="$(tr -d '[:space:]' < VERSION 2>/dev/null || echo "${VERSION}")" && \
    CGO_ENABLED=0 go build -trimpath \
    -ldflags "-X coin.local/coin-api/internal/version.Version=${VERSION}" \
    -o /coin-api ./cmd/coin-api

FROM ${REGISTRY}alpine:3.20

RUN apk add --no-cache ca-certificates && \
    addgroup -g 1000 app && adduser -u 1000 -G app -D app

COPY --from=build /coin-api /usr/local/bin/coin-api
COPY coin-api/openapi/v1.yaml /usr/share/coin/openapi/v1.yaml

RUN chown -R app:app /usr/share/coin/openapi

ENV COIN_OPENAPI_PATH=/usr/share/coin/openapi/v1.yaml

USER app

EXPOSE 8090
ENTRYPOINT ["coin-api"]
