FROM golang:1.22-bookworm AS base
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .

FROM base AS validate
RUN test -f go.mod && test -f main.go

FROM base AS test
RUN go test ./...

FROM base AS runtime
RUN CGO_ENABLED=0 go build -buildvcs=false -trimpath -o /out/app .

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=runtime /out/app /app/app
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
