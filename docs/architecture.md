# Архитектура Coin (Control Plane v2)

## Принцип разделения

```
┌─────────────────────────────────────────────────────────┐
│  coin-lib (Shared Library) — resolve, pod, creds, stages │
├─────────────────────────────────────────────────────────┤
│  coin-executor — validate, build engines, report        │
├─────────────────────────────────────────────────────────┤
│  coin-api + PostgreSQL + Nexus — manifest, GP, registry │
├─────────────────────────────────────────────────────────┤
│  coin-gp-content — build policy, Containerfile, schema  │
└─────────────────────────────────────────────────────────┘
```

Jenkins **не** checkout'ит platform content из git, **не** скачивает executor в bootstrap, **не** исполняет GP shell scripts.

## Компоненты

| Компонент | Назначение |
|-----------|------------|
| `coin-api` | Resolve manifest, registry metadata, GP admin |
| `coin-executor` | Runtime: validate, `run --stage`, publish, report |
| `coin-gp-content` | GP stacks: `content.yaml`, Containerfile, schema → Nexus |
| `coin-lib` | Jenkins glue only (`coinPipeline`) |
| `coin-agent` | Universal agent image (`coin-executor/Dockerfile.agent`) |
| `coin-starters` | Product scaffolding + thin `Jenkinsfile.coin` |
| `coin-ui` | Admin SPA |

**Удалено (superseded):** `coin-jenkins-agents/` — language stack images.

## Build engines

Политика сборки — в GP `content.yaml` → manifest `build.engine`:

| Engine | Sample GP | Кратко |
|--------|-----------|--------|
| `buildkit` | `go-app` | Multi-target Containerfile |
| `buildpack` | `go-app-bp` | Paketo `pack` + podman |
| `dockerfile` | `go-app-df` | Explicit Dockerfile targets |

Подробно: [agent-build-model.md](agent-build-model.md).

```
coin-executor run --stage validate  → schema + capabilities
coin-executor run --stage test      → engine-specific test
coin-executor run --stage build     → image → .coin/outputs.json
coin-executor run --stage publish   → registry push
```

## Продуктовый контракт

```yaml
coin:
  goldenPath: go-app
  version: "*"          # или =1.0.0, ~1.0.0, ^1.0.0

jenkins:
  credentials:
    docker: nexus-docker

project:
  name: my-service
  groupId: com.example.team
  repository: maven-releases
```

Strict v2 — поля `template` / `templateVersion` **не** поддерживаются.  
См. [config.md](config.md).

## GP composition (4 slots)

| Slot | Component | Manifest |
|------|-----------|----------|
| `agent` | `agent/coin-agent` | `runtime.image` |
| `executor` | `executor/coin-executor` | `executor.url` (baked в agent на pilot) |
| `lib` | `lib/coin-lib` | Jenkins `@Library` pin |
| `gp-content` | `gp-content/{gp}` | `build`, `pipeline`, `validateSchema` |

## Dynamic agent

Один K8s pod, один container — образ `manifest.runtime.image` (`coin-agent`).

Pod template рендерит `coin-lib` (`coinPodYaml`).

## Platform CI (local)

| Артефакт | Команда / job |
|----------|----------------|
| coin-agent | `make publish-agent` |
| coin-executor binary | `make coin-executor` / job `coin-executor` |
| gp-content | `make coin-gp-content` |
| coin-lib | `make coin-lib` |
| GP seed | `make seed-jenkins-lib` |
| E2E 3 engines | `make e2e-build-engines` |

## Связанные документы

- [control-plane.md](control-plane.md)
- [golden-paths.md](golden-paths.md)
- [jenkins-setup.md](jenkins-setup.md)
- [docker/README.md](../docker/README.md)
