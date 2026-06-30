## Context

Platform component lifecycle (spec `platform-component-lifecycle`) defines two states: `draft` → `published`. Delete draft описан для всех типов (`gp-content`, `branching-model`, `agent`), но реализован только для **GP releases** (`DELETE /v1/admin/golden-paths/{name}/versions/{version}`).

Runtime hub (`/platform/runtime/{profile}/releases`) показывает drafts после CI register, promote — вручную (`runtime-agent-registry`). Без delete операторы не могут убрать тестовые версии (например `agent-30-06@0.1.0-draft`).

Эталон реализации: `store.DeleteGPReleaseDraft` + `GpReleasesTab` / `GpReleaseDetail` delete UX.

## Goals / Non-Goals

**Goals:**

- Generic Admin API delete для component version draft (все platform types с draft lifecycle).
- Runtime UI: delete на Releases list + release detail для `type=agent`.
- Publisher-only; confirm dialog; redirect после delete.
- Audit `delete_component_draft`.

**Non-Goals:**

- Delete published versions.
- Delete component profile (`components` row).
- Nexus / registry cleanup.
- UI delete для build-stacks и branching-models (отдельный backlog).
- Block delete when version referenced in GP draft composition (v1: allow; promote/resolve fail as today).

## Decisions

### D1. API shape (mirror GP)

```
DELETE /v1/admin/components/{type}/{name}/versions/{version}?actor=
```

| Code | Условие |
|------|---------|
| 204 | draft deleted |
| 404 | version not found |
| 409 | `status != draft` |
| 401 | no admin key |

**Audit:** `delete_component_draft`, `entity_type=component_version`, `entity_key={type}/{name}@{version}`.

### D2. Store delete

Transaction:

1. `SELECT ... FOR UPDATE` component version; reject if not draft.
2. `DELETE FROM component_versions WHERE id = $1` — `component_artifact_bodies` cascade via FK.
3. Audit insert; commit.

**Не трогаем Nexus** (agent metadata-only; gp-content/bml drafts with Nexus package — delete PG row only, orphan Nexus blob acceptable per lifecycle spec).

### D3. UI placement (runtime agent)

| Surface | Action |
|---------|--------|
| `/platform/runtime/{name}/releases` | «Delete» link per draft row (publisher+) |
| `/platform/runtime/{name}/releases/{version}` | «Delete draft» button рядом с Publish (draft only) |

После delete: navigate to releases list. Confirm: `window.confirm` (как GP).

**Scope:** `compType === "agent"` only in this change; API ready for other families.

### D4. Permissions

Тот же gate что promote: `can("publisher")` + Admin API key на backend.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| GP draft ссылается на удалённый agent pin | Promote GP fail с blocking pin; acceptable v1 |
| CI re-register той же version после delete | 201 create draft снова (409 только concurrent duplicate) |
| build-stacks/branching без UI delete | API exists; UI backlog |

## Migration Plan

1. Deploy coin-api (backward compatible — new endpoint).
2. Deploy coin-ui.
3. No data migration.

Rollback: hide UI button; endpoint unused.

## Open Questions

| # | Вопрос | Статус | Решение |
|---|--------|--------|---------|
| Q1 | Block delete if pin in GP draft? | ✅ | v1: allow delete; GP promote gate catches stale pin |
