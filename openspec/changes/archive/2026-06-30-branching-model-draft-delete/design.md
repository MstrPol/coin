## Context

`runtime-agent-draft-delete` реализовал generic Admin API delete draft и UI для **agent** only. Branching models используют тот же platform hub pattern (`/platform/branching-models/{name}`), но draft UX отличается:

- Draft edit: `/platform/branching-models/{name}/{version}/edit` → `PlatformComponentEditor` + `BranchingModelEditor`
- Releases list: `/platform/branching-models/{name}/releases`
- Published detail: read-only release detail; drafts из hub ведут в editor (не flat metadata view как agent)

API `DELETE /v1/admin/components/branching-model/{name}/versions/{version}` уже работает (store cascade `component_artifact_bodies`, audit `delete_component_draft`, no Nexus).

Эталон UX: agent delete в `PlatformReleasesTab` + `PlatformComponentReleaseDetail`.

## Goals / Non-Goals

**Goals:**

- Branching-model hub: delete draft с Releases tab.
- Branching-model editor: delete draft в lifecycle panel (primary surface для draft work).
- Publisher-only; `window.confirm`; redirect на releases list после delete.
- E2E smoke + docs mention.

**Non-Goals:**

- coin-api changes.
- Delete published versions.
- Delete component profile.
- UI delete для `gp-content` (build-stacks backlog).
- Nexus cleanup.

## Decisions

### D1. Reuse API client as-is

`api.deleteComponentVersionDraft(compType, name, version, actor)` уже generic. Backend без изменений.

### D2. UI surfaces

| Surface | Component | Change |
|---------|-----------|--------|
| `/platform/branching-models/{name}/releases` | `PlatformReleasesTab` | Delete per draft row (как agent) |
| `/platform/branching-models/{name}/{version}/edit` | `PlatformComponentEditor` → `LifecyclePanel` | «Delete draft» рядом с Validate/Publish |
| Embedded editor on release detail | same `PlatformComponentEditor` | covered by D2 editor change |

**Решение:** заменить `isAgent`-only guard на helper `supportsDraftDelete(compType)` → `agent | branching-model`.

Альтернатива: отдельный branching-only component — отклонено (дублирование agent pattern).

### D3. Post-delete navigation

| From | Navigate to |
|------|-------------|
| Releases tab | reload list in place |
| Editor | `/platform/branching-models/{name}/releases` (через `familyHubPath`) |

### D4. Permissions

`can("publisher")` на UI; Admin API key на backend (как agent delete).

### D5. E2E

Extend `e2e-platform-component-hub.sh`: после create bml draft → `DELETE` → verify 404. Mirror agent delete block.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| GP draft pins deleted branching-model | Promote GP fails with blocking pin; v1 acceptable (как agent) |
| Unsaved editor changes on delete | Confirm dialog mentions irreversible delete |
| gp-content без delete UI | Explicit non-goal; API ready |

## Migration Plan

1. Deploy coin-ui only (API backward compatible).
2. No data migration.

Rollback: revert UI guards.

## Open Questions

| # | Вопрос | Статус | Решение |
|---|--------|--------|---------|
| Q1 | Block delete if pin in GP draft? | ✅ | v1: allow (same as agent) |
| Q2 | Delete из editor vs только Releases tab? | ✅ | Both — editor is primary draft surface |
