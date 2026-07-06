# Coin managed Containerfile: Go build targets for BuildKit engine.
# Не копируйте в репозитории сервисов.

FROM --platform=$TARGETPLATFORM golang:1.22-bookworm AS base
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .

FROM base AS validate
RUN test -f go.mod && test -f main.go

FROM base AS test
RUN go test ./...

FROM base AS artifact
RUN mkdir -p /out && CGO_ENABLED=0 go build -buildvcs=false -trimpath -o /out/app .

FROM --platform=$TARGETPLATFORM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=artifact /out/app /app/app
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
