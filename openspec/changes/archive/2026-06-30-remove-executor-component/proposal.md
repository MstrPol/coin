## Why

GP promote и resolve падают с `component version not found: executor/coin-executor@{agentVersion}`, хотя runtime уже полностью описан agent pin (`image` + `digest`, executor baked в образ). Отдельный platform component `executor` дублирует agent stack, требует ручной публикации парной версии и противоречит ADR [coin-ci-runtime](../../docs/adr/coin-ci-runtime.md) (bootstrap не качает binary).

Решение platform lead: **agent достаточно** — убрать `executor` как component и **полностью убрать** секцию `manifest.executor` (вариант B).

## What Changes

- **BREAKING:** Удалить component type `executor` из registry model (PG rows, admin register, GP validate derive, resolve augment).
- **BREAKING:** Удалить `manifest.executor` из resolved manifest v1 (`manifest.schema.json`, coin-api builder, coin-executor manifest struct, coin-lib glue).
- **BREAKING:** Убрать `derivedExecutorPin` из API и UI (agent release detail).
- GP validate/promote/resolver: только 3 pins (`agent`, `gp-content`, `branching-model`); никакого скрытого 4-го pin.
- `publish-executor.sh`: прекратить register в coin-api (Nexus upload для bake в agent image — по необходимости CI, вне registry).
- Deprecate/remove `coin-executor bootstrap` (download по `manifest.executor.url`).
- ADR `coin-ci-runtime.md`, docs, E2E — agent-only runtime model.

### Non-goals

- Переименование репозитория `coin-executor` (Go CLI остаётся).
- Изменение GP composition slots (остаётся 3 pins).
- Corp fleet rollout / migration wave.
- Build stack visual editor (отдельный change).
- Удаление `coin-branching-models/` git repo (отдельный change).

## Capabilities

### New Capabilities

_(нет — расширяем существующие спеки)_

### Modified Capabilities

- `runtime-agent-registry`: agent — единственный runtime component; убрать executor derive и manifest.executor.
- `platform-component-lifecycle`: убрать lifecycle правила для derived executor.
- `gp-composition-two-slot`: agent slot = полный CI runtime stack; resolve не materialize executor.
- `platform-component-hub`: убрать derived executor pin на agent release detail.
- `runtime-documentation`: three-pin + manifest без executor; supersede derive narrative.

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-api** | `platform_runtime.go`, `gp_release_prepare.go`, `store.go` augment, `manifest/builder.go`, schema, OpenAPI, component resolve modes |
| **coin-executor** | `internal/manifest`, `bootstrap` deprecate/remove, tests |
| **coin-lib** | `coinLoadConfig.groovy` — убрать executor layer |
| **coin-ui** | `PlatformComponentReleaseDetail`, `derivedExecutorPin` helper |
| **docker/e2e** | `e2e-platform-component-hub.sh`, seed scripts |
| **docs/adr** | `coin-ci-runtime.md`, `agent-build-model.md`, `config.md`, `control-plane.md` |
| **migrations** | Опционально: cleanup `components` type=executor (local pilot) |
