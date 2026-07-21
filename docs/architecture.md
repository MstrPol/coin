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
│  coin-api seed — bootstrap pipeline defaults            │
│  GP release — embedded pipeline authoring               │
└─────────────────────────────────────────────────────────┘
```

Jenkins **не** checkout'ит platform content из git, **не** скачивает executor в bootstrap, **не** исполняет GP shell scripts.

## Компоненты

| Компонент | Назначение |
|-----------|------------|
| `coin-api` | Resolve manifest, registry metadata, GP admin |
| `coin-executor` | Runtime: validate, `run --stage`, publish, report |
| `coin-lib` | Jenkins glue only (`coinPipeline`) |
| `coin-agent` | Universal agent image (`coin-executor/Dockerfile.agent`) |
| `coin-starters` | Product scaffolding + thin `Jenkinsfile.coin` |
| `coin-ui` | Admin SPA |

**Удалено (superseded):** `coin-jenkins-agents/` — language stack images.

## Build Stack vNext

Политика сборки и outputs — в GP/Build Stack → manifest `parameters`, `build.targets`, `deliverables`, `artifacts.containerfiles`, `pipeline` и `destinations`.

| Концепт | Назначение |
|---------|------------|
| `parameters` | Типизированные non-secret параметры stack |
| `build.targets[]` | Named targets; каждый target выбирает `buildkit` или `dockerfile` |
| `deliverables[]` | Named outputs с type, target reference и publish metadata |
| `artifacts.containerfiles[]` | Managed Containerfiles из gp-content package |
| `pipeline.stages[].steps[]` | Typed platform actions, исполняемые `coin-executor` |

Подробно: [adr/build-stack-vnext-contract.md](adr/build-stack-vnext-contract.md), [adr/coin-ci-runtime.md](adr/coin-ci-runtime.md), runbook [agent-build-model.md](agent-build-model.md).

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
  artifactId: my-service
```

Strict v2 — поля `template` / `templateVersion` **не** поддерживаются.  
См. [config.md](config.md).

Product config не содержит `project.repository`, `deliverables`, build/publish commands, pipeline stages, cache refs или registry/repository URLs. Эти решения принадлежат GP/Build Stack и материализуются в manifest.

## GP composition (three pins)

Оператор pin'ит в GP release composition:

| Slot | Component | Manifest |
|------|-----------|----------|
| `agent` | `agent/coin-agent` | `runtime.image`, `runtime.digest` |
| `gp-content` | `gp-content/{stack}` | `parameters`, `build.targets`, `pipeline`, `deliverables`, `artifacts`, `validateSchema` |
| `branching-model` | `branching-model/{name}` | `branching` |

Дополнительно GP version хранит `destinations` (`imageRegistryPrefix`, `buildCacheEnabled`, `artifactRepositoryBase`). Это поля версии GP, а не отдельный component pin.

**Не в GP composition map:** `lib` (platform pin для Jenkins `@Library`). Resolved manifest **не содержит** `executor`.

Jenkins credential IDs (например `nexus-docker`) не являются частью GP release или resolved manifest. Они задаются в product `.coin/config.yaml`, `coin-lib` defaults или Jenkins environment.

См. [adr/coin-ci-runtime.md](adr/coin-ci-runtime.md).

## Dynamic agent

Один K8s pod, один container — образ `manifest.runtime.image` (`coin-agent`).

Pod template рендерит `coin-lib` (`coinPodYaml`). Каноническая модель: [adr/coin-ci-runtime.md](adr/coin-ci-runtime.md).

## Platform CI (local)

| Артефакт | Команда / job |
|----------|----------------|
| coin-agent | `make publish-agent` |
| coin-executor CLI | Baked в `coin-agent` image; Nexus upload для bake (`publish-executor.sh`) |
| coin-lib | `make coin-lib` |
| GP seed | `make seed-jenkins-lib` (pipeline defaults из coin-api seed) |
| E2E 3 engines | `make e2e-build-engines` |

## Связанные документы

- [control-plane.md](control-plane.md)
- [golden-paths.md](golden-paths.md)
- [jenkins-setup.md](jenkins-setup.md)
- [docker/README.md](../docker/README.md)
