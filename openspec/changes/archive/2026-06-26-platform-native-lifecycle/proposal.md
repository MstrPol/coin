## Why

После рефакторинга IA раздел **Platform** (`/platform/runtime`, `/platform/build-stacks`, `/platform/branching-models`) стал естественной точкой управления компонентами, а **Component Studio** (`/studio`) — лишним indirection: каталоги ведут в Studio, operator теряет контекст типа компонента. Параллельно lifecycle `draft → canary → published` на уровне **компонентов** дублирует canary на уровне **GP** и усложняет модель без выигрыша для pilot fleet.

Сейчас: GP draft API принимает draft-компоненты, но UI composition editor показывает только `published`; promote GP не проверяет статусы pins. Модель не доведена до конца.

## What Changes

- **BREAKING**: Удалить маршрут `/studio` и nav-входы; type-aware editors встраиваются в entity pages под Platform.
- **BREAKING**: Убрать `canary` как статус `component_versions`; lifecycle компонентов: `draft` → `published` only.
- Canary pilot — **только на уровне GP**: `latest_canary` может указывать на GP draft; canary resolve допускает draft pins (build stack, branching-model).
- Runtime (agent/executor): **только `published`**, script-first publish; без draft UI.
- Build stacks (`gp-content`) и branching models: draft → publish через Platform entity pages (UI forms).
- GP draft может pin'ить `draft` + `published` компоненты; **GP promote** требует все pins = `published` (API gate + UI block).
- UI: статус каждого pin в composition; warning «этот pin — draft, может измениться» (canary = unstable by design, без lock).
- Amend ADR [gp-component-package-model.md](../../docs/adr/gp-component-package-model.md) и [canary.md](../../docs/canary.md).

### Non-goals

- Новый Backstage-like plugin или внешний developer portal.
- UI authoring для agent stack / Docker images (runtime остаётся script-first).
- Auto-lock draft компонентов при назначении на canary line.
- Wave rollout / corp fleet migration (corp gate).
- gp-content PG-only canary registry path (BML special case удаляется вместе с component canary).

## Capabilities

### New Capabilities

- `platform-component-lifecycle`: единый контракт draft/publish для platform components (gp-content, branching-model); runtime exception (published only); правила pin'ов в GP composition.

### Modified Capabilities

- `component-studio`: **REMOVED** — superseded by Platform entity pages.
- `platform-build-stacks`: in-place draft/create/edit/publish; убрать ссылки на Studio.
- `platform-runtime-catalog`: уточнить published-only, без draft UI.
- `branching-models-catalog`: in-place lifecycle; убрать Studio routing.
- `gp-publish-flows`: draft pins в composition picker; promote gate; draft-pin warnings.
- `gp-composition-two-slot`: resolve/pin rules по статусу компонента и каналу.
- `branching-model`: убрать component-level canary transitions; draft → published only.

### Related ADR (amend, not OpenSpec capability)

- `docs/adr/gp-component-package-model.md` — component canary superseded.
- `docs/canary.md` — GP draft на canary line разрешён.

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-api** | `ComponentResolveMode` refactor; promote GP validate pin statuses; migration `canary` → `published` или drop enum; remove component canary transitions |
| **coin-ui** | Platform entity pages с embedded editors; удалить `ComponentStudio`, `/studio` routes; composition picker + promote gate UI |
| **docs** | control-plane.md, golden-paths.md, canary.md, coin-ui-user-guide |
| **E2E** | обновить asserts на component status; seed без component canary |
| **OpenSpec** | sync 7 capability deltas; remove component-studio from main specs |
