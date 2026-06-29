# ADR: Coin CI Runtime

**Статус:** accepted (2026-06-16)  
**Operational SoT** для Jenkins CI pod, agent image, build engines и publish gate.

**Связанные ADR:** [build-engine-contract](build-engine-contract.md) (решение о введении `build.engine`), [gp-branching-model](gp-branching-model.md) (publish policy), [jenkins-lib-http-nexus](jenkins-lib-http-nexus.md) (coin-lib glue), [cicd-corp-migration-standards](cicd-corp-migration-standards.md) (corp target).

**Runbook:** [agent-build-model.md](../agent-build-model.md) (E2E, troubleshooting).

## Контекст

Coin CI runtime после hard cut:

- один universal **`coin-agent`** image вместо language-specific stack agents;
- build policy в **gp-content** → `manifest.build.engine`;
- orchestration в **coin-lib** (glue only), execution в **coin-executor**;
- GP composition — **three pins** (`agent`, `gp-content`, `branching-model`).

Этот ADR фиксирует **текущую** модель. Исторический контекст hard cut — в [build-engine-contract](build-engine-contract.md).

## Решение

### 1. Jenkins pod

```
┌──────────────────────────────────────────────────────────────┐
│ K8s pod: один container jnlp = manifest.runtime.image        │
│ (coin-agent)                                                  │
├──────────────────────────────────────────────────────────────┤
│ coin-lib: resolve → podTemplate → checkout → bootstrap       │
│ coin-executor: run --stage * → report                          │
└──────────────────────────────────────────────────────────────┘
```

| Инвариант | Значение |
|-----------|----------|
| Pod layout | Один container; **нет** dual pod (jnlp + stack) |
| Agent image | `coin-executor/Dockerfile.agent` → Nexus `coin-docker/coin-agent` |
| Host docker.sock | **Запрещён** |
| GP shell scripts | **Не** runtime path (`scripts/*.sh` superseded) |
| Executor bootstrap | Binary **baked** в agent; **не** curl в bootstrap |

### 2. Состав coin-agent image

| Компонент | Назначение |
|-----------|------------|
| `jenkins/inbound-agent` | JNLP remoting |
| `coin-executor` | validate, stages, publish, report |
| `podman` | Container builds; socket для `pack` |
| `pack` | Buildpack engine |
| `buildkitd` / `buildctl` | В образе; corp amd64 primary path |
| `paketo-builder.tar` | Baked builder (buildpack pilot) |

**Нет** language toolchains (Go/Java/Node) в agent — toolchain в managed Containerfile / builder images.

### 3. Bootstrap (coinPipeline)

Обязательно на каждой сборке:

1. `podman system service` → `unix:///var/run/docker.sock` **внутри pod** (не Docker Daemon).
2. **buildpack only:** `podman load` из `/usr/share/coin/paketo-builder.tar` при необходимости.
3. `coin-executor version`.

`buildkitd` **не** стартует в bootstrap на **local pilot arm64**.

### 4. Три build engine

Источник SoT: `coin-gp-content/stacks/<gp>/content.yaml` → manifest `build`.

| Engine | Sample GP | Containerfile | Реализация (pilot arm64) |
|--------|-----------|---------------|--------------------------|
| `buildkit` | `go-app` | multi-target | **podman build** по targets |
| `buildpack` | `go-app-bp` | нет | `pack build` + podman socket |
| `dockerfile` | `go-app-df` | explicit targets | **podman build** |

`coin-executor` dispatch по `manifest.build.engine`. `coin-lib` **не** интерпретирует engine.

Managed Containerfile materialize в workspace: `.coin/Containerfile` (content ref из gp-content package).

### 5. Pilot vs corp

| Environment | `buildkit` / `dockerfile` implementation | Bootstrap |
|-------------|------------------------------------------|-----------|
| **local pilot arm64** | podman build (buildctl RUN несовместим с nested runc в k3s) | podman system service only |
| **corp amd64** (roadmap) | buildkitd + buildctl | per [cicd-corp-migration-standards](cicd-corp-migration-standards.md) |

Имя engine `buildkit` в manifest **одинаково**; меняется implementation layer.

### 6. Typed pipeline stages

| Stage | Executor |
|-------|----------|
| `validate` | schema + capabilities |
| `test` | engine-specific test target / buildpack tests |
| `build` | image → `.coin/outputs.json` |
| `publish` | registry push |

Stages — typed ids; **нет** `pipeline.stages[].script.url`.

### 7. Publish gate (три слоя)

| Слой | Механизм |
|------|----------|
| 1. Jenkins | `params.publish=false` → coin-lib **skip** stage publish |
| 2. Jenkins → executor | `params.publish=true` → `COIN_PUBLISH_REQUEST=true` |
| 3. Executor | `manifest.branching` → deny publish с запрещённой ветки |

**Primary gate** — branching + Jenkins param. `pipeline.stages[].when: tag` **не** документируется как primary gate (legacy в reference `content.yaml`; cleanup в change `gp-content-schema-v2`).

См. [gp-branching-model](gp-branching-model.md), [how-to/branching-models.md](../how-to/branching-models.md).

### 8. GP composition (three pins)

Оператор pin'ит в GP release composition:

| Slot | Type | Manifest |
|------|------|----------|
| `agent` | `agent` | `runtime.image`, derived `executor` |
| `gp-content` | `gp-content` | `build`, `pipeline`, `validateSchema`, capabilities |
| `branching-model` | `branching-model` | `branching` |

| Не в GP composition | Где |
|---------------------|-----|
| `executor` | Derived от agent (`coin-agent@X` → `coin-executor@X`) |
| `lib` | Platform pin; Jenkins `@Library`; не в `gp_composition` map |

OpenSpec: `gp-composition-two-slot` (id retained; фактически three-pin).

### 9. Superseded (не реализовывать)

- `coin-jenkins-agents/`, language stack images
- `pipeline-bundle` component, orchestration bundle URL
- `manifest.jnlp`, dual-container pod
- GP `controls` в content.yaml как runtime contract (не wired)
- Host Docker Daemon, `scripts/*.sh` в runtime

## Последствия

- Документация runtime: этот ADR → [agent-build-model.md](../agent-build-model.md) → [architecture.md](../architecture.md).
- [build-engine-contract](build-engine-contract.md) — decision record о введении контракта; operational details — здесь.
- Reference `content.yaml` в `coin-gp-content` может содержать legacy поля до `gp-content-schema-v2`.

## Отклонённые альтернативы

- Language-specific agents — см. build-engine-contract.
- Project-level `build.engine` override — отдельный ADR.
- `controls` в gp-content как declarative Jenkins parameters — не реализовано; не часть runtime contract.
