## Context

**Аудитория:** enabling / platform team (operator console), не corp developer portal.

**Ограничение org:** архитекторы не допустят полноценный Backstage — дублирование с существующими целевыми системами. Coin UI — узкий **GP control plane admin**, интеграция с Jenkins/Gitea через links и build reports.

**Текущее:** `Layout.tsx` — horizontal nav, `max-w-5xl`; страницы Components, Branching Models, Studio, GP Releases разрозненно.

**Prerequisite:** `branching-model-lifecycle` ✅ archived; baseline specs в `openspec/specs/`.

## Goals / Non-Goals

**Goals:**

- Sidebar IA: Fleet | Golden Paths | Platform | Admin.
- Platform catalog по **ролям в composition**, не flat `component_type` list.
- gp-content: GP detail tab (primary) + Build stacks summary (secondary).
- Backstage-like patterns: catalog table + filters, entity → Studio link.
- Сохранить все существующие routes через redirects.

**Non-goals:**

- Backstage product/plugin.
- Новые coin-api endpoints.
- gp-content PG-only canary (BML follow-up).
- Project entity page full tabs в v1 (optional UI-4).

## Decisions

### D1: Два слоя UI

| Слой | Nav group | Сущности |
|------|-----------|----------|
| Fleet | Fleet | projects, build reports |
| Control plane | Golden Paths | GP releases, policy, resolve |
| Authoring | Platform | runtime, build stacks, branching, lib |
| Ops | Admin | platform settings, audit |

### D2: Platform taxonomy (не единый Components)

```
Platform
├── Runtime        agent/coin-agent, executor/coin-executor
├── Build stacks   gp-content/{gpName}  — 1:1 с GP profile name
├── Branching      branching-model/*   — shared policies
└── Jenkins lib    lib/coin-lib        — single-card catalog
```

### D3: gp-content placement (гибрид)

- **Primary:** GP detail → tab **Build stack** (versions, lifecycle, Open Studio).
- **Secondary:** Platform → **Build stacks** — сводная таблица всех gp-content.
- Rationale: enabling мыслит «evolve go-app», но platform lead сравнивает engines across GPs.

### D4: Studio entry

Studio не в top-level nav как peer. Доступ:
- Platform section → «Create» / entity row → Studio deep link.
- Optional pinned shortcut в sidebar footer для publisher ( `/studio` ).

### D5: Redirects (breaking nav only)

| Old | New |
|-----|-----|
| `/branching-models` | `/platform/branching-models` |
| `/components` | `/platform/components` (legacy, filter all types) |
| `/components/:type/:name` | keep or alias `/platform/...` |

### D6: Layout shell

- Left sidebar ~240px, collapsible on narrow viewports.
- Main content full width (убрать `max-w-5xl` constraint на shell level).
- Header: user, role, API docs link only.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Два пути к gp-content (GP tab vs Build stacks) | Same API data; tab primary, catalog secondary |
| Runtime pages script-first (no Studio) | Show versions + link to publish runbook |
| Large refactor breaks bookmarks | Redirects на old paths |
| Branching models spec conflict | Delta spec: nav under Platform |

## Open Questions

| # | Вопрос | Статус | Lean |
|---|--------|--------|------|
| Q1 | Studio shortcut в sidebar footer? | ⏳ | Да, publisher only |
| Q2 | Merge Canary + GP Policy в одну «Policy» page? | ⏳ | Phase 2; v1 keep routes, group in nav |
| Q3 | `/platform/components` legacy forever? | ⏳ | Deprecate label; hide from nav |
