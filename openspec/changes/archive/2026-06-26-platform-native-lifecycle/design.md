## Context

**Текущее состояние:**

- Platform IA (`/platform/runtime`, `/platform/build-stacks`, `/platform/branching-models`) — каталоги со ссылками в `/studio`.
- Component Studio — generic hub с type-aware editors для `gp-content` и `branching-model`.
- Component lifecycle: `draft` → `canary` → `published` (ADR [gp-component-package-model.md](../../docs/adr/gp-component-package-model.md)).
- GP lifecycle: `draft` → promote → `published`.
- Resolve modes: `stable` (published), `canary` (published+canary), `admin` (published+canary+draft).
- GP draft API (`ComponentResolveAdmin`) принимает draft pins; UI composition editor — только `published`.
- `docs/canary.md`: draft GP **не** попадает на canary line.

**Prerequisite:** `jenkins-lib-outside-platform` ✅ archived — lib вне control plane.

**Stakeholders:** enabling team (operator console).

## Goals / Non-Goals

**Goals:**

- Platform-native authoring: create/edit/publish компонентов in-place под `/platform/*`.
- Упростить lifecycle компонентов до `draft` → `published`.
- Canary pilot — **только GP-level**; canary line может указывать на GP draft.
- GP promote gate: все pins = `published` (agent всегда published).
- UI warnings для draft pins; canary = unstable by design.
- Полное удаление `/studio` route и nav.

**Non-goals:**

- UI authoring для agent stack / Docker images.
- Auto-lock draft при назначении на canary.
- Corp fleet rollout.
- Новые coin-api endpoints beyond lifecycle/refactor scope.

## Decisions

### D1: Убрать Component Studio как nav/route

**Решение:** удалить `/studio`, `/studio/:type/:name/:version`; editors переиспользовать как embedded pages.

**Альтернатива:** redirect alias — отклонено (полное удаление).

**Routes:**

```
/platform/build-stacks/:name/:version/edit     gp-content editor (draft only)
/platform/branching-models/:name/:version/edit branching-model editor (draft only)
/platform/build-stacks/:name/:version          detail (read-only if published)
/platform/branching-models/:name/:version      detail
```

### D2: Component lifecycle — draft → published only

**Решение:** убрать `canary` из `component_versions.status`. Переход: `draft` → validate → publish → `published` (immutable + Nexus package).

**Альтернатива:** оставить canary для BML PG-only path — отклонено; canary только на GP.

**Migration:** existing `canary` rows → `published` (если Nexus package есть) или остаются как `draft` + manual republish.

### D3: Runtime — published only, script-first

**Решение:** agent/executor не имеют draft UI. GP composition picker для agent — только `published` versions. Publish path — runbook/script.

### D4: Resolve matrix по каналу

| Channel / context | GP status | agent | gp-content, branching-model |
|-------------------|-----------|-------|----------------------------|
| Stable (`pin=*`, no canary) | published | published | published |
| Stable (exact pin) | published | published | published |
| Canary line | published **or draft** | published | draft + published |
| GP draft admin (create/update) | draft | published | draft + published |
| GP promote | draft→published | published | **published only** |

**Реализация:** refactor `ComponentResolveMode`:

- `ComponentResolveStable` — `published` only
- `ComponentResolveDraft` — `published` + `draft` (для GP draft edit и canary resolve)
- Agent/executor slots **всегда** `ComponentResolveStable` regardless of channel

Удалить `ComponentResolveCanary` (component canary status уходит).

### D5: Canary line на GP draft

**Решение (вариант A):** `catalog_policy.latest_canary` **может** указывать на GP version со `status=draft` (snapshot semver).

**Изменение docs/canary.md:** убрать «Draft/snapshot никогда не попадают в canary line».

Canary resolve для such GP uses `ComponentResolveDraft` for gp-content and branching-model slots.

### D6: GP promote gate

**Решение:** `PromoteDraftToPublished` MUST re-validate composition with `componentResolveModeForGPPublish` → all slots `ComponentResolveStable`.

API: HTTP 409 с списком `{type, name, version, status}` blockers.

UI: disabled Promote + CTA links to Platform entity publish flows.

### D7: Draft pin warnings (canary unstable)

**Решение:** UI показывает badge `draft` + warning text на composition rows и при назначении GP draft на canary line. **Без** lock/freeze draft components.

**Принято:** canary fleet может получать разный manifest при редактировании draft pin — by design.

### D8: Embedded editors

Переиспользовать editor logic из `ComponentStudio.tsx` (branching-model form, gp-content form) как child components в Platform entity pages. Не дублировать validation/publish API calls.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Migration `canary` component rows | One-time SQL migration + seed verification |
| Bookmarks `/studio/*` break | 404 or one-release redirect notice in changelog |
| Draft pin drift on canary fleet | UI warning; documented as unstable-by-design |
| Promote без re-validation сегодня | D6 — обязательный API gate в этом change |
| Large UI refactor | Reuse editor components; incremental: API first, then UI pages |

## Migration Plan

1. **coin-api migration:** map `component_versions.status='canary'` → `published` where Nexus package exists, else `draft`; drop canary from enum or stop writing it.
2. **coin-api:** refactor resolve modes; promote validation; remove component canary transition endpoints.
3. **coin-ui:** Platform entity pages; update composition picker; promote gate UI; remove Studio routes.
4. **docs:** amend ADR gp-component-package-model, rewrite canary.md.
5. **E2E:** update component status asserts; verify canary resolve with draft GP + draft pins on pilot project.

**Rollback:** revert migration Down if enum changed; UI rollback independent.

## Open Questions

| # | Вопрос | Статус | Решение |
|---|--------|--------|---------|
| Q1 | `canary` rows без Nexus package → draft или force publish? | ✅ | → `draft` (operator republish) |
| Q2 | `/studio/*` 404 vs redirect? | ✅ | 404 (полное удаление) |
| Q3 | GP draft на canary line | ✅ | `latest_canary` может = GP draft version |
