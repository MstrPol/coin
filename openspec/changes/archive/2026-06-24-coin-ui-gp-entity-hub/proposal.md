## Why

Golden Paths в coin-ui организованы **вокруг операций** (Releases, GP Policy, Canary), а не **вокруг сущности GP profile**. Страница `/releases` показывает плоский список всех версий всех GP — это вводит в заблуждение: enabling team ожидает каталог профилей (`go-app`), а не release train log.

Дополнительно: дублирование publish-потоков (`/releases/new-gp` сразу публикует `0.0.1`, `/releases/publish` — draft/publish/promote wizard), policy и canary «оторваны» от GP в отдельных nav-пунктах. После `coin-ui-enabling-ia` (Fleet/Platform entity pattern) Golden Paths — следующий логичный слой.

## What Changes

- **GP Catalog** (`/gp`) — список **профилей** (не версий): name, slots summary, latest stable/canary, release count.
- **GP Hub** (`/gp/:name`) — entity page с вкладками: Overview, Releases, Policy, Canary, Build stack.
- **Release detail** — `/gp/:name/releases/:version` (redirect с `/releases/:name/:version`).
- **Упростить publish IA:**
  - «Новый GP» — create profile; initial release — явный optional step (не скрытый auto-publish).
  - «New release» — primary action с GP hub (не глобальный `/releases/publish`).
  - Promote draft — только с release row / release detail.
- **Sidebar Golden Paths:** `GP Profiles` + `Resolve` (убрать Releases, GP Policy, Canary из top-level).
- **Redirects:** `/releases`, `/catalog`, `/canary` → соответствующие вкладки GP hub или catalog.

## Capabilities

### New Capabilities

- `gp-profile-catalog`: каталог GP profiles, summary columns, entry to GP hub.
- `gp-entity-hub`: GP hub tabs (Overview, Releases, Policy, Canary, Build stack), embedded бывшие Catalog/Canary/Releases list per GP.
- `gp-publish-flows`: унификация create profile / new release / promote без глобального publish wizard в nav.

### Modified Capabilities

- `ui-enabling-shell`: Golden Paths nav — GP Profiles + Resolve.
- `platform-build-stacks`: Build stack tab на GP hub (primary), release detail — optional link back.

## Impact

- **coin-ui:** новые pages/routes, refactor `GpReleases`, `Catalog`, `Canary`, `PublishWizard`, `CreateGPProfile`, `nav.ts`, docs.
- **coin-api:** без контрактных изменений (client-side join существующих endpoints).
- **Non-goals:** semver model, manifest format, corp fleet rollout, merge Resolve в GP hub.

## Non-goals

- Изменение GP release API или semver-модели.
- Удаление draft/snapshot workflow на backend.
- Global Resolve debug tool внутри GP hub (остаётся cross-GP).
- Backstage / dev portal.
