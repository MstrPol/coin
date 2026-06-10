# coin-executor

Runtime binary for Coin Control Plane. Bounded scope — see [CHARTER.md](CHARTER.md).

## Commands

```bash
coin-executor validate --project .coin/config.yaml --manifest .coin/manifest.json
coin-executor run --manifest .coin/manifest.json --stage validate   # или без --stage — все stages
coin-executor bootstrap --manifest .coin/manifest.json --dest ./coin-executor
coin-executor version
coin-executor report --manifest .coin/manifest.json --project .coin/config.yaml --build-url $BUILD_URL --result success
```

Env для legacy manifests с gitRef (deprecated): `COIN_CONTENT_DIR` / `COIN_PLATFORM_DIR`, `COIN_PLATFORM_GIT_URL`.

Platform-first (MVP-1+): stage scripts и Dockerfile скачиваются по `url` из manifest (Nexus). Git clone coin-platform **не нужен**.

## Dev

```bash
go test ./...
go build -o coin-executor ./cmd/coin-executor
```
