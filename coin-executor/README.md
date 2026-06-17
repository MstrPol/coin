# coin-executor

Runtime binary for Coin Control Plane. Bounded scope — see [CHARTER.md](CHARTER.md).

## Commands

```bash
coin-executor validate --project .coin/config.yaml --manifest .coin/manifest.json
coin-executor run --manifest .coin/manifest.json --stage validate
coin-executor run --manifest .coin/manifest.json --stage test
coin-executor run --manifest .coin/manifest.json --stage build
coin-executor run --manifest .coin/manifest.json --stage publish
coin-executor version
coin-executor report --manifest .coin/manifest.json --project .coin/config.yaml \
  --build-url "$BUILD_URL" --result success
```

`run` без `--stage` выполняет все stages из manifest по порядку.

## Build engines

Dispatch по `manifest.build.engine`:

| Engine | Реализация |
|--------|------------|
| `buildkit` | Multi-target Containerfile; local pilot arm64: **podman build** fallback |
| `buildpack` | `pack build` + podman socket |
| `dockerfile` | podman/buildkit по `imageTarget` / `testTarget` |

Подробно: [docs/agent-build-model.md](../docs/agent-build-model.md).

Content (Containerfile, schema) скачивается по URL из manifest → materialize в `.coin/`.

## coin-agent image

Universal Jenkins agent — [`Dockerfile.agent`](Dockerfile.agent).

Publish:

```bash
cd docker
make publish-agent GOARCH=arm64    # Apple Silicon
# или:
./scripts/publish-agent.sh 1.0.0 arm64
```

Содержит baked `coin-executor`, `podman`, `pack`, Paketo builder tar (buildpack).

## Dev

```bash
go test ./...
go build -o coin-executor ./cmd/coin-executor
```

Platform job: `make coin-executor` (Gitea + Jenkins).
