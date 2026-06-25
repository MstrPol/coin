## Why

После внедрения GP entity hub UI остаётся перегруженным и противоречит ментальной модели enabling team: профиль GP смешан с composition platform components, форма создания требует версии, которые никуда не сохраняются, в каталоге показывается бессмысленная колонка «Slots», а publish flow допускает direct release в обход draft. Нужно разделить **metadata профиля**, **GP-specific composition** (только в draft) и **platform runtime** (lib / agent stack).

## What Changes

- **BREAKING:** `gp_profiles` — только `name` + optional `description`; поле `slots` удаляется из контракта create/read profile (миграция существующих профилей).
- **BREAKING:** GP draft/release composition — **3 catalog pins:** `agent` (runtime stack), `branching-model`, `gp-content` (`agentStackName`, `branchingModelName`, `gpContentName`). Имя profile не привязано к component names.
- **BREAKING:** Platform runtime сужается до **`lib` only**; agent/executor выбираются в draft (executor bundled в agent stack per ADR).
- Operator UI: create profile — имя + description; hub — единственный CTA **New draft**; **direct publish** убран полностью; release только через promote draft; **draft можно удалить и редактировать composition**, published (в Nexus) — read-only, без delete.
- **BREAKING:** убрать вкладку **Build stack** с GP hub — gp-content фиксируется в composition **версии** GP; связи profile ↔ build stack нет; каталог компонентов — Platform → Build stacks.
- Каталог `/gp`: убрать колонку Slots.
- Fix: scoped draft wizard не показывает «Нет Golden Path» внутри существующего профиля.

## Non-Goals

- Fleet rollout / corp gate.
- Component Studio lifecycle (draft → canary → published) — без изменений.
- Удаление API `publishGPRelease` (может остаться для bootstrap/scripts; не в operator UI).
- Объединение `agent`+`executor` в один component type в registry (отдельный ADR при необходимости).

## Capabilities

### New Capabilities

- `gp-profile-metadata`: профиль GP как metadata-only сущность (`name`, optional `description`).
- `gp-composition-two-slot`: GP draft/release — **3 pins** (`agent`, `branching-model`, `gp-content`); capability id сохранён для delta traceability.
- `platform-runtime-line`: platform-level **`lib` only**; agent per GP draft.

### Modified Capabilities

- `gp-publish-flows`: draft-only operator path; удаление direct publish из UI.
- `gp-profile-catalog`: колонки каталога и форма create без slots/versions.
- `gp-entity-hub`: hub CTAs, overview empty state, welcome flow → first draft; убрать Build stack tab.
- `gp-entity-hub`: (same file) release detail — promote as sole publish path; composition incl. gp-content.
- `platform-build-stacks`: только Platform catalog; gp-content discovery с release composition, не с profile hub.

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-api** | `gp_profiles` schema/migration, `CreateGPProfile` API, composition validation, resolve materialization (platform runtime injection), OpenAPI |
| **coin-ui** | `CreateGPProfile`, `GpCatalogPage`, `GpHubLayout`, `PublishWizard`/draft pages, `GpOverviewTab`, удаление `/releases/new` |
| **docs** | `docs/golden-paths.md`, user guide, ADR cross-ref |
| **samples/E2E** | seed/bootstrap GP profiles, `demo-go-app` resolve path |

См. также ADR [build-engine-contract.md](../../docs/adr/build-engine-contract.md) (agent содержит executor), [jenkins-lib-http-nexus.md](../../docs/adr/jenkins-lib-http-nexus.md), [gp-component-package-model.md](../../docs/adr/gp-component-package-model.md).
