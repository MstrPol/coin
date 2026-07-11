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

Multi-stage сборка:
1. **executor-builder** (`golang:1.24-bookworm`) — `go build` бинаря `coin-executor`
2. **buildkit-bin** (`moby/buildkit`) — `buildkitd` / `buildctl` / `runc`
3. **runtime** (`jenkins/inbound-agent`) — podman + buildkit + baked binary

Publish вручную:

```bash
cd docker
make publish-agent VERSION=0.1.1 GOARCH=arm64
# или:
GOPROXY=... ./scripts/publish-agent.sh 0.1.1 arm64
```

Содержит baked `coin-executor`, `podman`, BuildKit (`buildkitd`/`buildctl`).

### Jenkins job `coin-executor`

Стадии: `Resolve version` → `Preflight` → `Test` → `Publish agent image`.

Job вызывает `scripts/publish-agent.sh`, который:
- собирает образ через `docker build -f Dockerfile.agent` (бинарь внутри multi-stage)
- пушит `coin-agent:<version>` в Nexus Docker
- регистрирует draft `agent/coin-agent@<version>` в coin-api
- **не** делает promote (manual gate в Platform UI)

Обязательные Jenkins credentials:
- `nexus-docker` (Username/Password)
- `coin-publisher-api-key` (Secret text)

Параметры job:
- `BUMP`: semver bump для `VERSION`
- `GOARCH`: `arm64` или `amd64` (target platform образа)
- `GOPROXY`: прокси для builder stage в Dockerfile

## Dev

```bash
go test ./...
go build -o coin-executor ./cmd/coin-executor
```

Platform job: `make coin-executor` (Gitea + Jenkins).
