## Why

Модель ветвления сегодня глобальна в `docs/branching.md` и не исполняется в runtime. Разные типы продуктов (сервисы vs библиотеки) требуют разных правил, но платформа не привязывает их к Golden Path.

Branching должен стать **5-м slot** GP composition — симметрично gp-content, executor, lib — с resolve → `manifest.branching` и runtime в coin-executor.

**Prerequisite:** [gp-component-platform](../archive/2026-06-23-gp-component-platform/) — ✅ archived.

## What Changes

- Новый component type **`branching-model`** (`trunk-based`, `semver-tag`)
- GP composition **4 → 5 slots** (`branching-model` pin в release)
- coin-api: load model artifact → `manifest.branching`
- coin-executor: `internal/branching` — ValidateBranch, ResolveVersion, ShouldPublish, Bump
- Component Studio UI-03: первый green field type (draft → canary → stable)
- **BREAKING:** все GP profiles/releases обновляются на 5 slots

## Capabilities

### New Capabilities

- `branching-model`: component type, model.yaml schema, Nexus package, resolve materializer
- `manifest-branching`: `manifest.branching` section, OpenAPI + manifest.schema.json
- `executor-branching`: runtime policy в coin-executor

### Modified Capabilities

- `component-platform`: 5-й composition slot (delta после GCP-0)

## Impact

- coin-api, coin-executor, coin-ui Component Studio
- GP profiles, seed scripts, demo-go-app E2E
- docs/branching.md, docs/golden-paths.md

## Non-goals

- Автоматический cherry-pick release → main
- Fleet migration wave
- Per-repo branching overrides (v1: без overrides)
- Jenkins multibranch include/exclude (остаётся Jenkins config)
