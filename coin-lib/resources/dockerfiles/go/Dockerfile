# Coin managed Dockerfile: Go runtime image.

FROM golang:{{GO_VERSION}}-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/app ./...

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app
ARG COIN_VERSION=0.0.0-local
LABEL org.opencontainers.image.version="${COIN_VERSION}"
ENV COIN_VERSION="${COIN_VERSION}"
COPY --from=builder /out/app /app/app
EXPOSE {{APP_PORT}}
CMD {{APP_CMD}}
