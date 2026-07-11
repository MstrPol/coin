## Why

Runtime (`agent`) в Platform зарегистрирован как component с draft lifecycle, но контракт metadata и UX не согласованы с реальным процессом: CI сегодня **auto-promote** после push, форма «New draft» смешивает version / image / digest / GOARCH без ясных двух путей (CI register vs ручной catch-up), а `goarch` не участвует в resolve и дублирует digest.

Нужен явный **image registry** контракт: profile = имя образа, version = semver тега, SoT = `image` + `digest`; promote — **только руками** на платформе как release gate.

Связи с branching / build-stacks нет — это отдельный слой, не YAML-редактор.

## What Changes

- **BREAKING:** CI (`publish-agent.sh`) регистрирует только **draft**; **убрать auto-promote** из скрипта.
- **BREAKING:** Promote `agent` версии — **только** через Platform UI / Admin API (publisher); coin-api **отклоняет** promote без `metadata.image` и `metadata.digest`.
- **BREAKING:** Удалить `goarch` из agent metadata (API, UI, OpenAPI, E2E fixtures); архитектура образа определяется digest.
- Уточнить agent metadata contract: `image` (runtime ref, tag = version), `digest` (`sha256:…`); инвариант tag ↔ version.
- Runtime UI: два явных пути — **CI draft** (read-only metadata + Promote) и **manual catch-up** (New draft с обязательными image + digest).
- **Убрать hardcoded switch** в `executorPinForAgentStack`: для **любого** agent profile derive `executor/coin-executor@{same version}`; pod runtime — только `image` + `digest` из agent metadata.
- Документация: amend `docs/adr/coin-ci-runtime.md`, `docs/agent-build-model.md`, runbook publish-agent; убрать упоминания GOARCH в platform metadata.

### Non-goals

- YAML/schema редактор для runtime (в отличие от build-stacks / branching).
- Изменения gp-content / branching-model редакторов.
- Удаление `/platform/components` — уже redirect на `/platform/runtime` (change `remove-platform-components-legacy` archived); повторная чистка только если найдутся хвосты в docs.
- Corp fleet rollout / HA coin-api.
- Auto-promote из Jenkins после digest verify (promote остаётся human gate).

## Capabilities

### New Capabilities

- `runtime-agent-registry`: контракт metadata agent, promote gate, CI register-only path, version↔tag invariant.

### Modified Capabilities

- `platform-component-lifecycle`: promote agent только publisher UI; digest required; убрать architecture из editable metadata; CI не promote.
- `platform-runtime-catalog`: UX двух путей; форма catch-up без GOARCH; promote CTA на release detail.
- `runtime-documentation`: ADR/docs отражают registry model, manual promote, без goarch в PG.

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-api** | Validate agent metadata on promote; reject missing digest; strip goarch from accepted fields |
| **coin-ui** | `PlatformNewDraftPage`, `PlatformAgentMetadataEditorPage`, release detail — упростить поля |
| **coin-executor** | `publish-agent.sh` — draft only, no promote |
| **docker/e2e** | `e2e-platform-component-hub.sh`, bootstrap — без auto-promote assumptions |
| **docs** | `agent-build-model.md`, ADR `coin-ci-runtime.md`, how-to publish agent |
| **OpenAPI** | agent metadata schema |
