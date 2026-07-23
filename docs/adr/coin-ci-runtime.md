# ADR: Coin CI Runtime

**Статус:** accepted (2026-06-16); **amended** 2026-07 — two-pin + embedded pipeline; **amended** 2026-07-23 — dual-container pod (`jnlp` + `builder`)  
**Operational SoT** для Jenkins CI pod, agent image, build engines и publish gate.

> **Amendment (2026-07):** GP composition — **2 pin** (`agent`, `branching-model`). Pipeline — embedded на GP release ([gp-embedded-pipeline](gp-embedded-pipeline.md)). Pin `gp-content` **удалён**. Исторические упоминания three-pin ниже — superseded.

> **Amendment (2026-07-23):** Pod layout — **два container**: официальный/корп `jnlp` + кастомный `builder` = `moby/buildkit:*-rootless` + baked `coin-executor`. Podman — **отдельный** service/sidecar (корп-паттерн), не bake в builder. Single-container `coin-agent` — **superseded**.

**Связанные ADR:** [build-engine-contract](build-engine-contract.md), [gp-branching-model](gp-branching-model.md), [jenkins-lib-http-nexus](jenkins-lib-http-nexus.md), [gp-embedded-pipeline](gp-embedded-pipeline.md).

**Runbook:** [agent-build-model.md](../agent-build-model.md). Layout: [workspace-layout.md](../workspace-layout.md).

## Контекст

Coin CI runtime (актуально):

- dual-container Jenkins pod: **`jnlp`** (официальный) + **`builder`** (кастомный);
- build/pipeline policy — **embedded GP release** → resolved manifest;
- orchestration в **coin-lib** (glue only), execution в **coin-executor** (baked в `builder`);
- GP composition — **two pins** (`agent`, `branching-model`).

## Решение

### 1. Jenkins pod

```
┌──────────────────────────────────────────────────────────────┐
│ K8s pod                                                       │
│  ┌─────────────────────┐  ┌────────────────────────────────┐ │
│  │ jnlp                │  │ builder                        │ │
│  │ jenkins/inbound-    │  │ moby/buildkit:*-rootless       │ │
│  │ agent (official/corp)│  │ + coin-executor (baked)        │ │
│  │ uid 1000 (jenkins)  │  │ uid 1000 (jenkins)             │ │
│  └─────────────────────┘  └────────────────────────────────┘ │
│ optional: podman system service (corp sidecar / DOCKER_HOST)  │
│ shared workspace (emptyDir / PVC)                             │
├──────────────────────────────────────────────────────────────┤
│ coin-lib: resolve → podTemplate → checkout (jnlp)            │
│ coin-executor: run --stage * → report  (container builder)   │
└──────────────────────────────────────────────────────────────┘
```

| Инвариант | Значение |
|-----------|----------|
| Pod layout | **Два** container: `jnlp` + `builder` |
| `jnlp` image | Официальный `jenkins/inbound-agent` (пин тега в coin-lib / defaults); **не** кастомный bake |
| `builder` image | `moby/buildkit:*-rootless` + baked `coin-executor` (`coin-executor/Dockerfile`) |
| RunAs | uid **1000** (`jenkins`) в `jnlp` и `builder` |
| BuildKit | **rootless** (base образа `builder`) |
| Podman | **Не** в образе `builder`; корп — отдельный service (`DOCKER_HOST=tcp://…`) |
| Host docker.sock | **Запрещён** |
| GP shell scripts | **Не** runtime path (`scripts/*.sh` superseded) |
| Executor bootstrap | Binary **baked** в `builder`; **не** curl в bootstrap |

### 2. Состав образов

#### `jnlp` (официальный)

| Компонент | Назначение |
|-----------|------------|
| `jenkins/inbound-agent` | JNLP remoting only |

Без Podman, BuildKit, `coin-executor`.

#### `builder` (кастомный = BuildKit rootless + bin)

| Компонент | Назначение |
|-----------|------------|
| `moby/buildkit:*-rootless` | rootless `buildkitd` / `buildctl` (корп: свой registry mirror + CA) |
| `coin-executor` | validate, stages, publish, report (baked bin) |

**Нет** Podman в этом образе. **Нет** language toolchains — toolchain в managed Containerfile / builder images.

> **Открытые вопросы:** optional podman sidecar в local pilot; corp `DOCKER_HOST` wiring; pin тега `inbound-agent` / корп agent.

### 3. Bootstrap (coinPipeline)

Обязательно на каждой сборке (**в container `builder`**):

1. `buildkitd` (если socket ещё нет) + `coin-executor version`.

Checkout / remoting — в `jnlp`. Stage `coin-executor run` — `container('builder')`.

Podman (если нужен) — внешний service; `DOCKER_HOST` задаёт platform/corp, не bootstrap builder.

### 4. Два build engine

Источник SoT: embedded pipeline GP release → resolved manifest (`build` / pipeline stages). Bootstrap defaults: `coin-api/internal/gpcontent/seed/`.


| Engine | Sample GP | Containerfile | Реализация (pilot arm64) |
|--------|-----------|---------------|--------------------------|
| `buildkit` | `go-app` | managed → `.coin/Containerfile` | **podman build** по targets |
| `dockerfile` (BYO) | `go-app-docker` | product `Dockerfile` | **podman build** по `imageTarget` / `testTarget` |

Buildpack superseded (hard cut 2026-06).

`coin-executor` dispatch по `manifest.build.engine`. `coin-lib` **не** интерпретирует engine.

Managed Containerfile materialize в workspace только для **buildkit** (из embedded GP pipeline / manifest content refs).

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
| `test` | engine-specific test target |
| `build` | image → `.coin/outputs.json` |
| `publish` | registry push |

Stages — typed ids; **нет** `pipeline.stages[].script.url`.

### 7. Publish gate (три слоя)

| Слой | Механизм |
|------|----------|
| 1. Jenkins | `params.publish=false` → coin-lib **skip** stage publish |
| 2. Jenkins → executor | `params.publish=true` → `COIN_PUBLISH_REQUEST=true` |
| 3. Executor | `manifest.branching` → deny publish с запрещённой ветки |

**Primary gate** — branching + Jenkins param. `pipeline.stages[].when: tag` **не** документируется как primary gate.

См. [gp-branching-model](gp-branching-model.md), [how-to/branching-models.md](../how-to/branching-models.md).

### 8. GP composition (two pins)

Оператор pin'ит в GP release composition:

| Slot | Type | Manifest |
|------|------|----------|
| `agent` | `agent` | `runtime.image`, `runtime.digest` (**builder**: BuildKit rootless + baked `coin-executor`) |
| `branching-model` | `branching-model` | `branching` |

`jnlp` image **не** pin в GP composition — фиксируется platform defaults (coin-lib).

Embedded pipeline на GP release → `pipeline` / related build fragments в manifest (**не** pin).

| Не в GP composition | Где |
|---------------------|-----|
| `lib` | Jenkins `@Library`; вне `gp_composition` |
| `gp-content` / `executor` | **удалены** как composition slots |

Resolved manifest v1 **не содержит** секцию `executor` — CI runtime полностью описан agent pin.

OpenSpec: `gp-release-two-pin`, `gp-composition-two-slot`, `gp-embedded-pipeline`.

### 10. Runtime agent registry (Platform)

| Поле metadata | Назначение |
|---------------|------------|
| `image` | Full container ref для Jenkins pod (`manifest.runtime.image`) |
| `digest` | Content-addressable pin (`sha256:…`); обязателен для promote |
| `runtime` | = `components.name` (profile) |

| Инвариант | Значение |
|-----------|----------|
| Promote | Только publisher (UI / Admin API); CI register — draft only |
| GOARCH | Build-time only (`publish-agent.sh`); **не** в PG metadata |
| Executor binary | Baked в **builder** image; **не** отдельный platform component |

### 9. Superseded (не реализовывать)

- `coin-jenkins-agents/`, language stack images
- `pipeline-bundle` component, orchestration bundle URL
- Single-container `coin-agent` (inbound-agent + tools + executor в одном образе)
- Fat custom `jnlp` image (Bake remoting + build tools вместе)
- GP `controls` в content.yaml как runtime contract (не wired)
- Host Docker Daemon, `scripts/*.sh` в runtime

## Последствия

- Документация runtime: этот ADR → [agent-build-model.md](../agent-build-model.md) → [architecture.md](../architecture.md).
- [build-engine-contract](build-engine-contract.md) — decision record о введении контракта; operational details — здесь.
- Seed pipeline YAML в `coin-api/internal/gpcontent/seed/` — bootstrap only; live SoT — GP release body / published manifest.

## Отклонённые альтернативы

- Language-specific agents — см. build-engine-contract.
- Project-level `build.engine` override — отдельный ADR.
- declarative Jenkins parameters из pipeline — вне этого ADR; не путать с product config.
