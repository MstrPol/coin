## Why

coin-ui перегружен: 11+ пунктов в horizontal nav, дублирование сущностей (Components, Branching Models, Studio) и нет ясного разделения **Fleet** (projects, builds) vs **Platform** (GP composition, runtime, build stacks, branching policy). Аудитория — **enabling/platform team**; полноценный Backstage в org недопустим (дублирование с существующими «целевыми» порталами). Нужен свой operator console с backstage-*подобными* паттернами (sidebar, catalog, entity pages), без fork Backstage.

## What Changes

- **BREAKING (nav):** horizontal top-nav → **sidebar** с группами; убрать отдельный top-level пункт `Branching Models` и generic `Components` как peer всему остальному.
- Новая IA **Fleet**: Overview, Projects (hub), Build reports.
- Новая IA **Golden Paths**: Releases, Policy (catalog+canary), Resolve (debug).
- Новая IA **Platform** (publisher+): Runtime (agent+executor), Build stacks (gp-content), Branching models, Jenkins library (coin-lib).
- **GP detail**: вкладка **Build stack** — primary path к gp-content версиям профиля.
- **Studio** — действие с entity pages, не отдельный «мир» в nav (остаётся shortcut для create).
- Role-gated sidebar: reader — Fleet + GP read-only; publisher — Platform + Studio.
- Redirects: `/branching-models` → `/platform/branching-models`; `/components` → catalog views с фильтром.

## Capabilities

### New Capabilities

- `ui-enabling-shell`: sidebar layout, nav groups, RBAC visibility, redirects со старых routes.
- `platform-runtime-catalog`: каталог agent + executor (versions, metadata, runbook links).
- `platform-build-stacks`: каталог gp-content + вкладка Build stack на GP detail.
- `fleet-project-hub`: project list с акцентом на GP pin / canary / last build (evolve existing Projects).

### Modified Capabilities

- `branching-models-catalog`: каталог под Platform, не отдельный top-nav peer.
- `component-studio`: deep links с platform entity pages и GP build stack tab.

## Impact

- **coin-ui**: `Layout`, routes, новые/переименованные pages, deprecate flat nav.
- **coin-api**: без контрактных изменений (client-side join существующих endpoints).
- **docs**: `coin-ui-user-guide.md`, `coin-ui/README.md` nav tables.
- **Non-goals:** Backstage embed/plugin; dev self-service portal; gp-content PG-only lifecycle (follow-up); corp fleet rollout.

## Non-goals

- Внедрение Spotify Backstage или plugin ecosystem.
- TechDocs, scaffolder, org-wide software catalog.
- Product developer audience как primary users.
- API/OpenAPI changes.
- gp-content lifecycle alignment с BML (отдельный change).
- Полный project entity page с tabs (можно phased в UI-4).
