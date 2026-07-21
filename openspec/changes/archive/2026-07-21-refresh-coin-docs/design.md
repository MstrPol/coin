## Context

Фактический канон (код + OpenSpec):

| Тема | Спека / ADR |
|------|-------------|
| GP composition | `gp-release-two-pin`, `gp-composition-two-slot` — pins `agent` + `branching-model` |
| Pipeline | `gp-embedded-pipeline` — body на GP release; seed `coin-api/internal/gpcontent/seed/` |
| Jenkins | `jenkins-lib-boundary` — coin-lib glue only |
| Runtime | `coin-ci-runtime` ADR + agent image |

Документы в `docs/` частично обновлены после remove branching-models / gp-content, но **массово** ещё содержат three-pin, Studio, gp-content package, неточный seed.

**Workspace layout сейчас (факт):**

```
coin-workspace/                    # интеграционный workspace (не один git обязательно)
├── coin-api/                      # sibling — control plane API
├── coin-executor/                 # sibling — CLI + coin-agent image
├── coin-lib/                      # sibling — Jenkins Shared Library
├── coin-ui/                       # sibling — Admin SPA
├── coin/                          # meta: docs, openspec, docker pilot, starters, samples
│   ├── docs/
│   ├── openspec/
│   ├── docker/                    # local compose (Gitea, Jenkins, Nexus, …)
│   ├── coin-starters/
│   └── samples/
└── samples/                       # может дублировать / зеркалить E2E (уточнить при apply)
```

Corp P4-03 — отдельный шаг: каждый sibling → свой Gitea `coin/<name>`; `coin/` meta не обязан становиться одним prod app-repo.

Спека `runtime-documentation` **сама устарела** (требует three-pin) — блокирует честный docs↔specs sync.

## Goals / Non-Goals

**Goals:**

1. Docs narrative = OpenSpec (2-pin + embedded pipeline + Platform hubs).
2. Одна каноническая страница layout: workspace today + corp split.
3. Починить `runtime-documentation` spec.
4. Индекс и cross-links без 404 на удалённые папки.

**Non-Goals:** corp split execution; UI code cleanup; pipeline v4; archive rewrite.

## Decisions

### D1. SoT для narrative = `openspec/specs/`, не «что написано в старом ADR»

При конфликте ADR vs spec: обновляем **docs + ADR banner**, requirements — из **активных** specs (`gp-release-two-pin`, `gp-embedded-pipeline`). Исторические ADR сохраняют superseded/amendment, не выдаются за live contract.

### D2. Новый doc: `docs/workspace-layout.md` (+ усилить `prod-repo-split.md`)

| | |
|--|--|
| **Выбор** | Новый top-level `docs/workspace-layout.md` — «как устроен workspace сейчас»; `runbooks/prod-repo-split.md` — «как режем в corp». Architecture ссылается на оба. |
| **Почему** | Смешивать «уже sibling checkouts» и «corp filter-repo» в одном runbook путает local pilot. |
| **Альтернатива** | Всё в prod-repo-split — corp banner отпугивает; layout не находят. |

Содержание `workspace-layout.md` (обязательно):

1. Диаграмма sibling repos vs `coin/` meta.
2. Таблица: path → роль → local make/CI → будущий corp repo.
3. Removed: `coin-gp-content`, `coin-branching-models` (fixtures/seed elsewhere), `coin-jenkins-agents`.
4. Seed SoT: branching → `docker/testdata/`; pipeline → `coin-api/.../seed/`.
5. Pointer на P4-03 / corp gate.

### D3. Fix `runtime-documentation` в том же change

MODIFIED: three-pin → two-pin + embedded pipeline; Purpose update; scenarios architecture/control-plane.

### D4. Порядок правок docs (при apply)

1. Specs (runtime-documentation + new docs-monorepo-layout + gp-composition-two-slot delta).
2. `workspace-layout.md` + `prod-repo-split.md`.
3. `architecture.md`, `control-plane.md`, `golden-paths.md`, `coin-ci-runtime.md` banner.
4. How-to / jenkins / responsibilities / agent-build / README.
5. ADR superseded banners где live text врёт.
6. Grep-аудит: `three-pin`, `gp-content` pin, `coin-gp-content/`, `/studio`, `make coin-gp-content`, `Component Studio` как primary.

### D5. Spec inventory mapping (для apply)

| Doc | Primary specs |
|-----|---------------|
| architecture, control-plane | gp-release-two-pin, gp-embedded-pipeline, jenkins-lib-boundary, runtime-documentation |
| golden-paths, publish-gp-release | gp-publish-flows, gp-entity-hub, gp-embedded-pipeline |
| Platform UI mentions | platform-*-catalog, ui-enabling-shell, branching-models-catalog |
| Build engines | build-engine, coin-ci-runtime ADR |
| Workspace / split | docs-monorepo-layout |

## Risks / Trade-offs

| Риск | Митигация |
|------|-----------|
| Объём docs → огромный PR | Части по D4; tasks чеклист по файлам |
| `samples/` в двух местах | Зафиксировать в workspace-layout при apply (какой канон) |
| platform-build-stacks spec ещё жив, UI мёртв | Docs: Build stacks **removed** per ADR embedded; UI cleanup — non-goal |
| runtime-documentation vs coin-ci-runtime ADR drift | Обновить оба согласованно |

## Migration Plan

1. Delta specs → sync main specs.
2. Написать `workspace-layout.md`, обновить prod-repo-split.
3. Refresh top-level + how-to.
4. ADR banners.
5. `rg` audit запрещённых live-терминов (кроме superseded секций).
6. Обновить `docs/README.md` reading order.

**Rollback:** git revert docs/specs; поведение runtime не меняется.

## Open Questions

| # | Вопрос | Статус | Варианты | Решение |
|---|--------|--------|----------|---------|
| Q1 | Канон `samples/` — `coin/samples` или workspace `samples/`? | ✅ | A: coin/samples | `coin/samples` — пишет `docker/scripts/samples.sh` |
| — | BLOCKING архитектурных решений нет | — | — | 2-pin уже accepted |
