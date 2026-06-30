## Why

После `runtime-agent-registry` CI регистрирует agent drafts (`publish-agent.sh` → draft only), а promote — ручной gate. Orphan drafts (тестовые профили вроде `agent-30-06`, неудачные CI register) копятся в hub Releases без способа убрать их из UI.

Спека `platform-component-lifecycle` и hub уже требуют delete draft для platform components, но **API и runtime UI не реализованы** (есть только `DELETE` для GP releases). Нужно закрыть gap до возврата к build-stacks work.

## What Changes

- **coin-api:** `DELETE /v1/admin/components/{type}/{name}/versions/{version}` — только `status = draft`; 409 для published; audit `delete_component_draft`; cascade `component_artifact_bodies` (agent — обычно пусто).
- **coin-ui (runtime):** «Delete draft» на Releases tab (`/platform/runtime/{name}/releases`) и release detail для agent drafts (publisher+).
- **coin-ui api client:** `deleteComponentVersionDraft(type, name, version)`.
- OpenAPI + E2E smoke для agent draft delete.
- Docs: краткое упоминание в `docs/agent-build-model.md` / user guide (cleanup orphan drafts).

### Non-goals

- Delete published agent versions (immutable).
- Delete agent **profile** (component row) — только version draft.
- UI delete для build-stacks / branching-models в этом change (API generic; UI — runtime first).
- Удаление Docker image из Nexus при delete draft.
- Corp fleet rollout.

## Capabilities

### New Capabilities

_(нет — закрываем gap в существующих спеках)_

### Modified Capabilities

- `platform-component-lifecycle`: явный Admin API delete draft + audit action.
- `platform-runtime-catalog`: delete draft на Releases tab и release detail для agent.
- `platform-component-hub`: уточнить release detail actions (agent delete wired).

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-api** | store delete, admin handler, route, OpenAPI |
| **coin-ui** | `PlatformReleasesTab`, `PlatformComponentReleaseDetail`, `api.ts` |
| **docker/e2e** | optional step в `e2e-platform-component-hub.sh` |
| **docs** | agent publish / UI guide — cleanup drafts |
