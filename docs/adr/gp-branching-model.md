# ADR: GP Branching Model (5-slot composition)

**Статус:** accepted (GBM-0.1, 2026-06-23)  
**Контекст:** [gp-branching-model](../../openspec/changes/gp-branching-model/)  
**Prerequisite:** [gp-component-package-model](gp-component-package-model.md) (archived ✅)  
**Связанные ADR:** [control-plane-v2.md](control-plane-v2.md), [build-engine-contract.md](build-engine-contract.md)

## Контекст

Модель ветвления описана в [docs/branching.md](../branching.md) как глобальный документ, но **не исполняется** в runtime: `coin-executor` не читает правила из manifest, версия продукта не привязана к GP pin.

Разные типы продуктов (сервисы vs библиотеки) требуют разных правил. Платформа должна pin'ить модель ветвления на уровне **Golden Path release**, симметрично `gp-content`, `lib`, `executor`, `agent`.

Component Studio уже поддерживает authoring `branching-model` (GCP-1 green field). Следующий шаг — 5-й composition slot и runtime в coin-executor.

## Решение

### 1. Component type `branching-model`

| Поле | Значение |
|------|----------|
| Type | `branching-model` |
| Primary artifact | `model.yaml` |
| Authoring | Component Studio → Nexus package → registry |
| Reference catalog | `coin-branching-models/models/<name>/` (human docs + export) |

Первая пара моделей local pilot:

| Model | GP profiles | Sample |
|-------|-------------|--------|
| `trunk-based` | `go-app`, `go-app-bp`, `go-app-df` | `demo-go-app*` |
| `semver-tag` | `go-lib`, `java-maven-app` (future) | unit/demo |

### 2. GP composition: 4 → 5 slots

**BREAKING** для новых GP releases после GBM rollout:

| Slot | Type | Manifest section |
|------|------|------------------|
| `agent` | `agent` | `runtime` |
| `executor` | `executor` | `executor` |
| `lib` | `lib` | `lib` |
| `gp-content` | `gp-content` | `build`, `pipeline`, `validateSchema`, … |
| **`branching-model`** | **`branching-model`** | **`branching`** |

Profile задаёт **имя** модели (`trunk-based`); GP release pin'ит **версию** (`1.0.0`).

`ValidateCanonicalGPSlots` расширяется с 4 до 5 слотов. Существующие 4-slot releases остаются readable до re-publish с 5-м слотом (compat window — один GP semver bump per profile).

### 3. Manifest section `branching`

Materializer загружает `model.yaml` из Nexus package (или `content_ref` v2 manifest subset) и эмитит:

```json
{
  "branching": {
    "name": "trunk-based",
    "version": "1.0.0",
    "trunk": { "branch": "main" },
    "branchTypes": ["feature", "bugfix", "release"],
    "versioning": { "tagPrefix": "v", "qualifiers": { "snapshot": { "enabled": true }, "rc": { "enabled": true } } },
    "publish": { "when": "tag" }
  }
}
```

Полная структура — subset из `model.yaml` (см. Component Studio `buildManifestSubset`). Executor **не** ходит в coin-api за моделью во время build — только manifest snapshot.

### 4. Runtime ownership: coin-executor

Вся policy logic — в `coin-executor/internal/branching/`:

| API | Назначение |
|-----|------------|
| `ValidateBranch` | имя ветки vs `branchTypes` / patterns |
| `ResolveVersion` | `COIN_VERSION` из git tags + rules |
| `ShouldPublish` | publish stage eligibility |
| `Bump` | `coin version bump` (future CLI) |

**Инварианты:**

- `run --stage publish` **не обходит** branching policy (fix bypass).
- Docker image tag = `COIN_VERSION` из `ResolveVersion`, **не** `goldenPath.version` pin.
- Jenkins `params.publish=true` — manual override только platform/debug; продуктовый default — policy из manifest.

`coin-lib` **не** интерпретирует branching (coin-lib-scope).

### 5. Schema

- `branching-model.schema.json` — валидация `model.yaml` при Studio validate/register.
- `manifest.schema.json` + OpenAPI — секция `branching`.

### 6. UI-first lifecycle

Тот же lifecycle, что у других platform components:

`draft` → validate → register → `canary` → pilot health gate → `published`.

Promote wizard (GCP-2) применим к `branching-model` versions.

## Superseded

| Документ / решение | Замена |
|--------------------|--------|
| Глобальная модель только в `docs/branching.md` | GP-pinned `manifest.branching` + каталог моделей |
| 4-slot GP composition | 5-slot с `branching-model` |
| Implicit trunk-based для всех GP | Explicit pin per GP profile |

## Non-goals (v1)

- Per-repo branching overrides
- Автоматический cherry-pick release → main
- Fleet wave migration (corp gate)
- Jenkins multibranch include/exclude (остаётся Jenkins config)

## Открытые вопросы

| # | Вопрос | Статус | Решение |
|---|--------|--------|---------|
| Q1 | Approve 5-й slot + breaking composition | ✅ GBM-0.3 | 5-slot + legacy 4-slot compat |
| Q2 | Per-repo overrides v1 | ✅ | none |
| Q3 | `params.publish` override | ✅ | platform/debug only, зафиксировать в runbook |

## Последствия

**Плюсы:** предсказуемая product version; разные GP — разные модели; единый Studio path.

**Минусы:** breaking GP publish; миграция seed + E2E; executor scope расширяется (в рамках ADR).

## Rollout (local pilot)

1. GBM-0: ADR + 5-slot profile API
2. GBM-1: manifest materializer + `coin-branching-models` catalog
3. GBM-2: executor branching package
4. GBM-3: seed 5-slot + `demo-go-app` E2E publish policy
5. GBM-4: docs index (`branching.md` → catalog)

См. [openspec/changes/gp-branching-model/tasks.md](../../openspec/changes/gp-branching-model/tasks.md).
