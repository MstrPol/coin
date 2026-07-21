# ADR: GP Component Package Model (UI-first)

**Статус:** accepted (GCP-0.1, 2026-06-23); **amended** platform-native-lifecycle (2026-06-16)  
**Контекст:** [gp-component-platform](../../openspec/changes/gp-component-platform/)  
**Связанные ADR:** [control-plane-v2.md](control-plane-v2.md), [jenkins-lib-http-nexus.md](jenkins-lib-http-nexus.md), [build-engine-contract.md](build-engine-contract.md)

> **Amendment (2026-06):** component-level `canary` **superseded**. Lifecycle: `draft` → `published` only. Canary rollout — **GP-level** (`catalog.latest_canary`); draft pins на canary line разрешены для `gp-content` / `branching-model`. Component Studio (`/studio`) удалён в пользу Platform entity routes.

> **Superseded (2026-07):** platform component type **`gp-content`** удалён. Pipeline-inline v3 — embedded body GP release. Composition — 2 pin (`agent`, `branching-model`). См. [gp-embedded-pipeline.md](gp-embedded-pipeline.md).

## Контекст

Golden Path release собирается из platform components (`agent`, `executor`, `gp-content`, `branching-model`). Jenkins Shared Library (`coin-lib`) **не** является platform component — см. [jenkins-lib-outside-platform.md](jenkins-lib-outside-platform.md). Сегодня каждый тип прошёл свой путь публикации:

- разные shell-скрипты и форматы `metadata` / `content_ref`;
- dual-write артефактов в PostgreSQL (`component_artifact_bodies`, `gp_artifact_bodies`) и Nexus;
- authoring platform content через git/Gitea + Jenkins jobs;
- `coin-ui` публикует GP composition, но не редактирует content компонентов;
- `loadComposition` в coin-api — switch per component type.

Enabling team не имеет единого playbook. Принято направление **UI-first**: draft → canary на pilot GP → promote stable ([docs/planning.md](../planning.md)).

## Решение

### 1. UI-first authoring

Enabling team создаёт и выпускает platform components **только** через coin-ui Component Studio → Admin API → Nexus.

| Роль | Путь |
|------|------|
| Primary | coin-ui → coin-api Admin API → Nexus package → registry |
| Optional | git export/import для review, bulk edit, audit snapshot |
| Deprecate | `make coin-*`, Gitea platform jobs, shell `publish-*.sh` как primary path |

Gitea на local pilot остаётся для **product samples** и optional dev mirror — не для platform component release.

### 2. Component lifecycle

**Superseded:** значение `canary` в `component_status` и переход `draft`→`canary`→`published`.

| State | Product resolve (stable) | Canary channel (GP) | Platform draft edit |
|-------|--------------------------|---------------------|---------------------|
| `draft` | ❌ | ✅ (draft pins) | ✅ |
| `published` | ✅ | ✅ | ❌ |

**Draft SoT:** artifact bodies + `content_ref` manifest subset в PostgreSQL (без Nexus `package.url`).

**Published SoT:** immutable Nexus package + `content_ref` v2 с `package.url` (CI fallback).

Переход: `draft` → validate → register (PG) → **promote** (Nexus) → `published`.

Интеграция с canary GP ([docs/canary.md](../canary.md)): `catalog.latest` / `latest_canary` (может указывать на GP `draft`), `project.canary_mode`, `X-Coin-Channel`.

GP promote gate: все composition pins должны быть `published` (API 409 + `blockingPins`).

Promote catalog wizard: `catalog.latest_canary`→`latest` + component promote (для legacy pilot flows).

### 3. Component Package Model

**Nexus layout (immutable):**

```
maven-releases/coin/{type}/{name}/{version}/
  package.manifest.json    # files, sha256, roles
  package.zip              # optional bundle
  ... artifact files ...
```

**PostgreSQL `component_versions`:** только ссылки — package URL, digest, typed `content_ref` v2. Без дублирования больших тел для **новых** releases.

**content_ref v2 (форма):**

```json
{
  "package": {
    "url": "http://nexus/.../package.manifest.json",
    "sha256": "sha256:..."
  },
  "manifest": {
    "build": {},
    "pipeline": { "stages": [] }
  }
}
```

Для простых types (`branching-model`) поле `manifest` = parsed model rules.

### 4. Generic resolve materializers

coin-api resolve **не** знает деталей типа напрямую:

1. Registry component types → materializer
2. Composition slot → load package → materialize manifest section(s)
3. Итоговый product manifest — денормализованный snapshot всех slots

CI fallback — **только Nexus** (manifest blob + component packages), не PG bodies.

### 5. Composition slots (текущее и целевое)

| Slot | Type | Manifest section | Статус |
|------|------|------------------|--------|
| agent | agent | `runtime` | as-is |
| executor | executor | `executor` | as-is |
| lib | lib | `lib` (новое) | GCP-4 |
| gp-content | gp-content | `build`, `pipeline`, … | migrate GCP-3 |
| branching-model | branching-model | `branching` | GBM-0.2 ✅ |

## Инвентаризация (as-is → to-be)

| Component | Authoring | Platform CI | PG registry | PG bodies | Nexus | В manifest | Consumer |
|-----------|-----------|-------------|-------------|-----------|-------|------------|----------|
| **gp-content** | ~~git `coin-gp-content`~~ **superseded** — embedded GP pipeline + `coin-api/.../seed` | ~~`publish-content.sh`~~ | — | — | — | via GP manifest | coin-executor |
| **executor** | git | `publish-executor.sh` | ✅ metadata.url | — | maven binary | ✅ executor | coin-agent |
| **agent** | Dockerfile.agent | `publish-agent.sh` | ✅ metadata.image | — | Docker registry | ✅ runtime.image | Jenkins pod |
| **lib** | git → Gitea tag | `coin-lib.sh` | partial | — | target ZIP | ✅ `lib` section (GCP-4) | Jenkins `@Library` + manifest zip ref |
| **GP manifest** | — | resolve side-effect | manifest_hash/url | gp_artifact_bodies copy | blob + pointers | whole JSON | coin-lib |
| **seed (legacy)** | embedded bytes | bootstrap | draft | gp_artifact_bodies | — | — | deprecate |

**To-be (все types):** Component Studio → package → Nexus → register → composition pin → resolve.

## Deprecations

| Что | Когда | Замена |
|-----|-------|--------|
| `publish-*.sh` как primary | GCP-3+ | Admin API + Studio |
| Dual-write `gp_artifact_bodies` на GP release | GCP-5 | Nexus-only snapshot — [runbook](../runbooks/gp-artifact-bodies-migration.md) |
| `component_artifact_bodies` для published | GCP-5 | Nexus package (draft PG — см. Q1) |
| Embedded `gpcontent/seed` bytes | GCP-5 | UI seed / `seed-jenkins-lib` — [runbook](../runbooks/gp-artifact-bodies-migration.md) Phase C |
| Gitea platform component jobs | GCP-4 | Nexus HTTP + Studio |
| Switch per type в `loadComposition` | GCP-3 | Materializer registry |
| `buildControls` в metadata без manifest | GCP-3 | content_ref v2 manifest subset |

## Открытые вопросы (✅ platform lead — 2026-06-23)

| # | Вопрос | Решение |
|---|--------|---------|
| Q1 | `component_artifact_bodies` | **PG для draft/canary Studio;** published — Nexus package + optional PG preview |
| Q2 | lib в manifest | **section `lib` + zip ref** (GCP-4) |
| Q3 | agent storage | **PG metadata + Docker registry** |
| Q4 | Порядок work | **GCP-0+1 параллельно branching green field** |

**Ранее закрыто:** Q5 UI-first + optional git export; Q6 Gitea только product samples.

## Отклонённые альтернативы

| Альтернатива | Почему нет |
|--------------|------------|
| Git-only platform components | не масштабируется для enabling team; нет единого control plane UX |
| UI-only без export | нет PR review path для platform engineers |
| Продолжать ad-hoc publish per type | увеличивает хаос (риск при branching-model) |
| Dual path git + UI release | противоречит hard cut инвариантам Coin |

## Последствия

**Плюсы:**

- один playbook для enabling team;
- предсказуемый resolve и Nexus fallback;
- canary lifecycle для components синхронен с GP canary;
- новые component types без нового switch в коде.

**Минусы:**

- миграция gp-content и lib — отдельные фазы;
- миграция `component_status` + API;
- platform lead должен закрыть Q1–Q4 до GCP-1 API implementation.

## Критерии приёмки ADR

- [x] UI-first lifecycle зафиксирован
- [x] Package model + content_ref v2 описаны
- [x] Inventory + deprecations в одном документе
- [x] Q1–Q4 подтверждены platform lead (2026-06-23)
