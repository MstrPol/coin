## 1. Routes & shell

- [x] 1.1 `GpHubLayout` + path tabs: `/gp/:name`, `/gp/:name/releases`, `policy`, `canary`, `build-stack`
- [x] 1.2 `GpCatalogPage` at `/gp` — profile rows (join gpNames + catalog + release counts)
- [x] 1.3 Redirects: `/releases`, `/catalog`, `/canary`, `/releases/:n/:v`
- [x] 1.4 Update `nav.ts`: Golden Paths → GP Profiles + Resolve

## 2. GP hub tabs (reuse components)

- [x] 2.1 Extract `Catalog` body → `GpPolicyTab` (props: gpName)
- [x] 2.2 Extract `Canary` body → `GpCanaryTab`
- [x] 2.3 Extract release table from `GpReleases` → `GpReleasesTab` (single GP)
- [x] 2.4 Move Build stack from `GpReleaseDetail` → `GpBuildStackTab` on hub
- [x] 2.5 `GpOverviewTab` — slots, latest pins, quick links

## 3. Publish flows

- [x] 3.1 `CreateGPProfile`: profile only + optional initial release CTA on hub
- [x] 3.2 GP-scoped «New release» route (from PublishWizard publish tab)
- [x] 3.3 GP-scoped «New draft» on Releases tab
- [x] 3.4 Remove global Publish from nav; redirect `/releases/publish`
- [x] 3.5 Release detail at `/gp/:name/releases/:version`; promote draft only there + release row

## 4. Cleanup & docs

- [x] 4.1 Deprecate or thin `GpReleases`, `Catalog`, `Canary` standalone pages
- [x] 4.2 Update `coin-ui/README.md`, `docs/coin-ui-user-guide.md`
- [x] 4.3 Manual smoke: catalog → hub tabs → new profile → new release → redirect bookmarks
- [x] 4.4 `openspec validate coin-ui-gp-entity-hub --strict`
