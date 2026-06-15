FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=0.1.0-dev
RUN VERSION="$(tr -d '[:space:]' < VERSION 2>/dev/null || echo "${VERSION}")" && \
    CGO_ENABLED=0 go build -trimpath \
    -ldflags "-X coin.local/coin-api/internal/version.Version=${VERSION}" \
    -o /coin-api ./cmd/coin-api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /coin-api /usr/local/bin/coin-api
COPY openapi/v1.yaml /usr/share/coin/openapi/v1.yaml
ENV COIN_OPENAPI_PATH=/usr/share/coin/openapi/v1.yaml
EXPOSE 8090
ENTRYPOINT ["coin-api"]
