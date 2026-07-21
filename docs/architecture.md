# Архитектура Coin (Control Plane v2)

Канон требований: `openspec/specs/` (`gp-release-two-pin`, `gp-embedded-pipeline`, `jenkins-lib-boundary`, `runtime-documentation`).  
Расположение репозиториев: [workspace-layout.md](workspace-layout.md).

## Принцип разделения

```
┌─────────────────────────────────────────────────────────┐
│  coin-lib (Shared Library) — resolve, pod, creds, stages │
├─────────────────────────────────────────────────────────┤
│  coin-executor (в coin-agent) — validate, run, publish   │
├─────────────────────────────────────────────────────────┤
│  coin-api + PostgreSQL + Nexus — manifest, GP, registry  │
├─────────────────────────────────────────────────────────┤
│  GP release — embedded pipeline (authoring)              │
│  coin-api seed — bootstrap pipeline defaults             │
└─────────────────────────────────────────────────────────┘
```

Jenkins **не** checkout'ит platform content из git, **не** скачивает executor binary в bootstrap, **не** исполняет GP shell scripts из product repo.

## Компоненты

| Компонент | Назначение |
|-----------|------------|
| `coin-api` | Resolve manifest, registry, GP admin, bootstrap seed |
| `coin-executor` | Runtime: `validate`, `run --stage` / `--task`, `publish`, `report` |
| `coin-lib` | Jenkins glue only (`coinPipeline`) |
| `coin-agent` | Universal agent image (`coin-executor/Dockerfile.agent`) |
| `coin-starters` | Product scaffolding + thin `Jenkinsfile.coin` |
| `coin-ui` | Admin / Platform SPA |

**Удалено:** `coin-jenkins-agents/`, `coin-gp-content/`, `coin-branching-models/` — см. [workspace-layout](workspace-layout.md).

## GP composition (two pins)

Оператор pin'ит в GP release composition **ровно два** внешних слота:

| Slot | Component | Manifest |
|------|-----------|----------|
| `agent` | `agent/coin-agent` | `runtime.image`, `runtime.digest` |
| `branching-model` | `branching-model/{name}` | `branching` |

**Embedded pipeline** (parameters, stages/steps, containerfiles, validateSchema) — body GP release, **не** composition pin.  
**Не в composition:** `gp-content`, `executor`, `lib`.

Дополнительно GP version хранит `destinations` (`imageRegistryPrefix`, `buildCacheEnabled`, `artifactRepositoryBase`).

См. [adr/gp-embedded-pipeline.md](adr/gp-embedded-pipeline.md), [adr/coin-ci-runtime.md](adr/coin-ci-runtime.md).

## Pipeline и engines

Политика сборки задаётся **embedded pipeline** на GP release и материализуется в resolved manifest (`parameters`, `pipeline`, `build` / deliverables по актуальному schema). Runtime исполняет `coin-executor`.

```
coin-executor run --stage validate
coin-executor run --stage test
coin-executor run --stage build
coin-executor run --stage publish
```

Engines (local pilot): `buildkit`, BYO `dockerfile`. Buildpack — superseded.  
Runbook: [agent-build-model.md](agent-build-model.md).

## Продуктовый контракт

```yaml
coin:
  goldenPath: go-app
  version: "1.0.0"    # или *, =1.0.0, ~1.0.0, ^1.0.0
```

Strict v2 — поля `template` / `templateVersion` **не** поддерживаются. См. [config.md](config.md).

## Dynamic agent

Один K8s pod, один container — `manifest.runtime.image` (`coin-agent`).  
Pod template рендерит `coin-lib`. Канон: [adr/coin-ci-runtime.md](adr/coin-ci-runtime.md).

## Platform CI (local)

| Артефакт | Команда |
|----------|---------|
| coin-agent | `make publish-agent` |
| coin-lib | `make coin-lib` / `coin-lib-http` |
| GP + branching seed | `make seed-jenkins-lib` |
| E2E | `make e2e-build-engines` / demo jobs |

## Связанные документы

- [workspace-layout.md](workspace-layout.md)
- [control-plane.md](control-plane.md)
- [golden-paths.md](golden-paths.md)
- [jenkins-setup.md](jenkins-setup.md)
- [docker/README.md](../docker/README.md)
