# syntax=docker/dockerfile:1.8
ARG REGISTRY=""

FROM ${REGISTRY}golang:1.25-alpine AS build
WORKDIR /src

# Копируем go.mod и go.sum и загружаем зависимости с использованием кеша
ARG GOPROXY="https://proxy.golang.org,direct"
ARG GOPRIVATE=""
ARG GOINSECURE="*"
ARG GIT_SSL_NO_VERIFY=true

ENV GOINSECURE=${GOINSECURE} \
    GIT_SSL_NO_VERIFY=${GIT_SSL_NO_VERIFY}

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=secret,id=go_auth \
    if [ -f /run/secrets/go_auth ]; then cat /run/secrets/go_auth > ~/.netrc; fi && \
    go mod download

# Копируем исходный код
COPY . .

# Сборка бинарника с использованием кеша сборки и модулей
ARG VERSION=0.1.0-dev
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    VERSION="$(tr -d '[:space:]' < VERSION 2>/dev/null || echo "${VERSION}")" && \
    CGO_ENABLED=0 go build -trimpath \
    -ldflags "-X coin.local/coin-api/internal/version.Version=${VERSION}" \
    -o /coin-api ./cmd/coin-api

# Финальный образ
FROM ${REGISTRY}alpine:3.20

# Устанавливаем сертификаты и создаем non-root пользователя (id 1000)
RUN apk add --no-cache ca-certificates && \
    addgroup -g 1000 app && adduser -u 1000 -G app -D app

# Копируем артефакты
COPY --from=build /coin-api /usr/local/bin/coin-api
COPY openapi/v1.yaml /usr/share/coin/openapi/v1.yaml

# Настройка прав для non-root пользователя
RUN chown -R app:app /usr/share/coin/openapi

ENV COIN_OPENAPI_PATH=/usr/share/coin/openapi/v1.yaml

# Запуск от имени non-root
USER app

EXPOSE 8090
ENTRYPOINT ["coin-api"]
