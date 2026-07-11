## 1. Open questions (blocking before API pin rules change)

- [x] 1.1 Зафиксировать Q1 (draft agent pins в GP draft): default A — published-only в composition; обновить design Open Questions → ✅
- [x] 1.2 Зафиксировать Q2 (auto-promote после Jenkins): выбор A/B; обновить runbook

## 2. coin-api — agent draft lifecycle

- [x] 2.1 Allow `CreateDraftComponentVersion` for `type=agent` without content_ref validation
- [x] 2.2 Agent promote: `PromoteComponentToPublished` без Nexus package requirement
- [x] 2.3 Optional: `derivedExecutorPin` в GET component version response для agent
- [x] 2.4 Unit tests: agent draft create, promote, immutable published metadata
- [x] 2.5 OpenAPI: document agent draft register + promote; idempotent 409 on duplicate draft

## 3. coin-executor — publish-agent migration

- [x] 3.1 `publish-agent.sh`: `POST .../versions/drafts` вместо direct publish
- [x] 3.2 Добавить promote step per Q2 decision (CI auto или documented manual)
- [x] 3.3 Обновить runbook в `docs/` для agent publish flow

## 4. coin-ui — shared hub shell

- [x] 4.1 `PlatformComponentHubLayout` (family, profile label, tabs, «New draft»)
- [x] 4.2 `PlatformProfileCatalogPage` — profile-grouped list, «New profile»
- [x] 4.3 Overview + Releases tab components (reuse patterns from `GpOverviewTab` / `GpReleasesTab`)
- [x] 4.4 `PlatformNewProfilePage` для трёх families (или shared form)
- [x] 4.5 `PlatformNewDraftPage` — create draft version under hub

## 5. coin-ui — runtime (agent) hub

- [x] 5.1 Refactor `PlatformRuntimePage` → profile catalog
- [x] 5.2 Agent hub routes в `App.tsx`: `/platform/runtime/:name`, releases, detail
- [x] 5.3 Agent release detail: metadata + derived executor read-only line
- [x] 5.4 Agent draft metadata edit form (image, digest, goarch catch-up)
- [x] 5.5 Promote / delete draft actions on agent release detail

## 6. coin-ui — build-stacks hub

- [x] 6.1 Refactor `PlatformBuildStacksPage` → profile catalog
- [x] 6.2 Hub routes: `/platform/build-stacks/:name`, releases tab, release detail
- [x] 6.3 Перенести create flow: catalog «New profile», hub «New draft»
- [x] 6.4 Redirect `/platform/build-stacks/:name/:version` → `.../releases/:version`
- [x] 6.5 Унифицировать label «Create draft» → «New draft»

## 7. coin-ui — branching-models hub

- [x] 7.1 Refactor `BranchingModelsPage` → profile catalog с «New profile»
- [x] 7.2 Hub routes: `/platform/branching-models/:name`, releases, detail
- [x] 7.3 Hub «New draft» + publish/delete на release detail
- [x] 7.4 Redirect flat URLs → hub hierarchy

## 8. coin-ui — GP composition links

- [x] 8.1 `GpReleaseDetail`: composition links → platform hub `/releases/{version}` URLs
- [x] 8.2 Legacy redirect `/components/agent/:name` → `/platform/runtime/:name`

## 9. E2E и validation

- [x] 9.1 E2E: platform hub navigation (runtime, build-stacks, branching-models)
- [x] 9.2 E2E: agent draft register + promote flow
- [x] 9.3 E2E: create branching model profile + first draft from hub
- [x] 9.4 `openspec validate platform-component-hub --strict`
- [x] 9.5 Обновить `coin-ui/README.md` и `docs/coin-ui-user-guide.md`
