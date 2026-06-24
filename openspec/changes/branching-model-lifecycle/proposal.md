## Why

Enabling team не видит branching models как отдельный продуктовый объект: в общем списке Components недостаточно статусов, версий, pilot/canary контекста и быстрых действий. Параллельно Studio при **publish canary** вызывает `register-package`, который **сразу пишет в Nexus** — это противоречит ADR Q1 (`component_artifact_bodies` только для draft/canary в PG; Nexus — при stable) и договорённости: canary проверяется через resolve API на данных из БД, а immutable package в Nexus появляется только после promote **published** (fallback для CI без live DB).

## What Changes

- Новый раздел UI **Branching Models** (`/branching-models`): каталог всех моделей, версии, статусы (`draft` / `canary` / `published`), привязка к GP profiles, ссылки в Studio и promote wizard.
- **BREAKING (lifecycle):** `register-package` / publish-to-canary для Studio-компонентов **не загружает** артефакты в Nexus; сохраняет `content_ref` v2 с **PG-only** manifest subset (без `package.url` до promote).
- Promote to **published** — единственная точка Nexus upload + финальный `content_ref` с `package.url` / `sha256` (immutable fallback).
- Resolve **canary channel** читает branching-model из PG (`component_artifact_bodies` + content_ref subset) без Nexus.
- Resolve **stable channel** для `published` — Nexus package (как сегодня); PG bodies опционально для admin preview.
- Обновить Component Studio flow: Validate → Publish canary (PG + status) → pilot health → Promote stable (Nexus).
- Bootstrap-скрипт `publish-branching-model.sh` — перевести на тот же контракт или пометить deprecated для fleet (local seed может вызывать promote-with-nexus admin path).

## Capabilities

### New Capabilities

- `branching-models-catalog`: UI-каталог branching models с фильтрами по статусу, версиям и GP usage.

### Modified Capabilities

- `component-studio`: canary publish без Nexus; Nexus только на promote to published.
- `component-platform`: уточнение component lifecycle — draft/canary SoT в PG, Nexus immutable только на published.
- `branching-model`: resolve rules для canary без Nexus package (PG materializer).

## Impact

- **coin-ui**: новая страница/меню, доработка Studio publish/promote UX.
- **coin-api**: `RegisterComponentPackage`, `PromoteComponentToPublished`, resolve materializers для canary/draft.
- **OpenAPI** + Admin API контракт `content_ref` v2 (optional `package` до published).
- **docs/adr/gp-component-package-model.md**: amend Q1 lifecycle (canary без Nexus).
- **coin-branching-models/scripts/publish-branching-model.sh**: align с promote-only Nexus.
- **Non-goals:** fleet migration wave; per-repo overrides; изменение Jenkins multibranch; полный рефактор gp-content lifecycle в этом change (только общий механизм + branching-model first).

## Non-goals

- Corp fleet rollout и массовая миграция существующих Nexus packages.
- Отдельный lifecycle для gp-content в этом change (follow-up, тот же паттерн).
- Автоматический cherry-pick release → main.
- Jenkins-side branching logic (coin-lib-scope).
