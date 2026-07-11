## Context

**Ментальная модель enabling team:** «evolve go-app» — один GP profile, внутри него releases, policy, canary, build stack.

**SoT (уже в coin-api):**

| Слой | API | Роль |
|------|-----|------|
| Profile | `gp_profiles`, `GET .../profile` | слоты composition |
| Releases | `gp_releases`, `GET .../golden-paths` | версии manifest |
| Policy | `GET/PATCH .../catalog` | latest, minimum, deprecated |
| Canary | `GET/PATCH .../canary` | rollout %, health |

**Текущий UI (проблемы):**

```
/releases          → flat table ALL versions (misleading as "GP list")
/releases/new-gp   → createGPProfile + publish 0.0.1 (hidden combo)
/releases/publish  → 3-tab wizard (draft | publish | promote)
/catalog           → policy per GP (dropdown)
/canary            → canary per GP (dropdown)
```

Prerequisite: `coin-ui-enabling-ia` ✅, `platform-build-stacks` tab on release detail.

## Goals / Non-Goals

**Goals:**

- Entity-centric Golden Paths IA (как Platform catalog).
- GP catalog как landing page раздела.
- Policy + Canary + Releases — вкладки на GP hub, не peer nav items.
- Ясные publish entry points без дублирования.
- Bookmarks: redirects со старых routes.

**Non-goals:**

- coin-api changes.
- Объединение Policy и Canary в одну вкладку (разная семантика — две вкладки на hub).
- Удаление Resolve из nav.

## Decisions

### D1: Route map

| New | Replaces |
|-----|----------|
| `GET /gp` | `/releases` (catalog intent) |
| `GET /gp/:name` | hub default tab Overview |
| `GET /gp/:name/releases` | filtered release list (was `/releases?name=`) |
| `GET /gp/:name/policy` | `/catalog` + gp selector |
| `GET /gp/:name/canary` | `/canary` + gp selector |
| `GET /gp/:name/build-stack` | build stack (from platform-build-stacks) |
| `GET /gp/:name/releases/:version` | `/releases/:name/:version` |

Hub uses **path tabs** (`/gp/go-app/policy`) для bookmarkable state.

### D2: GP Catalog columns

| Column | Source |
|--------|--------|
| Name | `gpNames` |
| Slots | `gpProfile.slots` count / keys |
| Latest stable | `catalog.latest` |
| Latest canary | `catalog.latestCanary` |
| Releases | count published (+ badge drafts) |
| Actions | Open hub |

Client-side join: `gpNames` + parallel `catalog` + `gpReleases(name)` per row (pilot scale ≤20 GPs ok; follow-up batch API if needed).

### D3: GP Hub layout

```
┌─────────────────────────────────────────────────┐
│ go-app                          [New release]   │
│ 5 slots · latest 1.0.0 · canary 1.1.0-snapshot  │
├─────────────────────────────────────────────────┤
│ Overview │ Releases │ Policy │ Canary │ Build  │
├─────────────────────────────────────────────────┤
│ <tab content — reuse existing page components>  │
└─────────────────────────────────────────────────┘
```

Reuse: extract tab bodies from `Catalog.tsx`, `Canary.tsx`, release table from `GpReleases.tsx`, build stack from `GpReleaseDetail` tab.

### D4: Publish flow consolidation

| Scenario | Entry | Behavior |
|----------|-------|----------|
| Bootstrap new GP | GP catalog → «New profile» | `createGPProfile` only; optional step 2 «Publish initial release» |
| New semver release | GP hub → «New release» | Linear flow (ex-publish tab «publish») scoped to `:name` |
| Draft snapshot | GP hub Releases → «New draft» | ex-publish tab «draft» |
| Promote draft | Release row / release detail only | Remove promote tab from global wizard |

**Deprecate:** top-nav link to `/releases/publish`. Route remains + redirect to `/gp/:name/releases/new` when `?name=` present.

### D5: Sidebar

```
Golden Paths
├── GP Profiles   /gp
└── Resolve       /resolve
```

### D6: Redirects

| Old | New |
|-----|-----|
| `/releases` | `/gp` |
| `/releases/:n/:v` | `/gp/:n/releases/:v` |
| `/catalog` | `/gp` (or `/gp/go-app/policy` if `?name=` — optional) |
| `/catalog?name=go-app` | `/gp/go-app/policy` |
| `/canary?name=go-app` | `/gp/go-app/canary` |
| `/releases/publish?name=go-app` | `/gp/go-app/releases/new` |

### D7: New profile — decouple auto-publish

`CreateGPProfile` today:

```ts
await api.createGPProfile(...);
await api.publishGPRelease(trimmed, { version: "0.0.1", ... });
```

**Change:** after create → navigate to `/gp/:name` with banner «Profile created. Publish initial release?» + CTA. Initial release version configurable (default `0.0.1`).

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| N+1 API calls on GP catalog | Pilot ≤20 GPs; batch endpoint — follow-up |
| Large refactor of PublishWizard | Phase: embed as GP-scoped routes first |
| Broken bookmarks | Redirects table above |
| Duplicate Build stack (release detail vs hub) | Hub primary; release detail link «Open build stack» |

## Open Questions

| # | Вопрос | Статус | Lean |
|---|--------|--------|------|
| Q1 | Initial release optional after new profile? | ⏳ | Да, explicit CTA |
| Q2 | GP catalog default sort | ⏳ | name ASC |
| Q3 | Keep `/promote` global page? | ⏳ | Deprecate; promote on release only |
