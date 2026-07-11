## Why

После `runtime-agent-draft-delete` generic Admin API `DELETE /v1/admin/components/{type}/{name}/versions/{version}` уже удаляет draft для всех platform types, включая `branching-model`. UI delete подключён только для **agent** (`isAgent` guard в hub Releases и release detail).

Orphan branching-model drafts (эксперименты в rule builder, неудачные версии до publish) остаются в hub без способа убрать их из UI. Спека `platform-component-lifecycle` и `platform-component-hub` требуют delete draft для всех типов с draft lifecycle — для branching models gap не закрыт.

## What Changes

- **coin-ui (branching-models):** «Delete» / «Delete draft» на Releases tab (`/platform/branching-models/{name}/releases`) для draft rows.
- **coin-ui editor:** «Delete draft» в lifecycle panel редактора (`/platform/branching-models/{name}/{version}/edit` и embedded editor на release detail).
- Переиспользовать существующий `api.deleteComponentVersionDraft("branching-model", …)` — без изменений coin-api.
- **E2E:** шаг delete branching-model draft в `e2e-platform-component-hub.sh` (после create draft).
- **Docs:** краткое упоминание cleanup drafts в `docs/how-to/branching-models.md` или `docs/coin-ui-user-guide.md`.

### Non-goals

- Изменения coin-api / OpenAPI (endpoint уже реализован).
- Delete published branching-model versions.
- Delete branching-model **profile** (component row) — только version draft.
- UI delete для build-stacks (`gp-content`) — отдельный backlog.
- Удаление Nexus package при delete draft (PG row + artifact bodies cascade; orphan Nexus blob acceptable).
- Block delete when version pinned in GP draft (v1: allow; promote gate catches stale pin).

## Capabilities

### New Capabilities

_(нет)_

### Modified Capabilities

- `branching-models-catalog`: delete draft actions на hub Releases и в editor lifecycle panel.
- `platform-component-hub`: уточнить draft release detail / editor actions для `branching-model` delete.
- `platform-component-lifecycle`: сценарий delete branching-model draft через существующий Admin API (symmetry с agent).

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-ui** | `PlatformReleasesTab`, `PlatformComponentEditor` (+ optional `PlatformComponentReleaseDetail`) |
| **docker/e2e** | `e2e-platform-component-hub.sh` — bml draft delete |
| **docs** | branching models how-to / UI guide — cleanup drafts |
| **coin-api** | без изменений (reuse `DeleteComponentVersionDraft`) |
