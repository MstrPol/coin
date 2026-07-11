## 1. ADR и документация (blocking)

- [x] 1.1 Amend `docs/adr/gp-component-package-model.md` — component canary superseded; draft→published only; canary только GP-level
- [x] 1.2 Переписать `docs/canary.md` — GP draft на canary line разрешён; draft pins unstable by design
- [x] 1.3 Обновить `docs/control-plane.md`, `docs/golden-paths.md` — Platform-native lifecycle, без Studio
- [x] 1.4 Обновить `docs/coin-ui-user-guide.md` — новые Platform routes, promote gate, draft warnings

## 2. coin-api — migration и resolve modes

- [x] 2.1 Migration: `component_versions.status='canary'` → `published` (если Nexus package) или `draft`; stop writing canary
- [x] 2.2 Refactor `ComponentResolveMode`: `Stable` + `Draft`; удалить `Canary`; agent slots always Stable
- [x] 2.3 Обновить resolve service: canary channel + GP draft → `ComponentResolveDraft` для gp-content/branching-model
- [x] 2.4 Canary line: разрешить `latest_canary` → GP `status=draft`
- [x] 2.5 Unit tests: resolve matrix (stable/canary/GP draft); component_resolve_test update

## 3. coin-api — component lifecycle

- [x] 3.1 Удалить component canary transitions (`publish_component_canary`, promote component canary)
- [x] 3.2 Упростить publish flow: draft → validate → published (Nexus upload) для gp-content и branching-model
- [x] 3.3 Удалить `usesPGOnlyCanaryRegistry` и BML special-case paths
- [x] 3.4 OpenAPI: убрать component canary schemas/endpoints; document promote gate error shape

## 4. coin-api — GP promote gate

- [x] 4.1 `PromoteDraftToPublished`: re-validate composition — all pins `published`
- [x] 4.2 HTTP 409 response с payload blocking pins `{type, name, version, status}`
- [x] 4.3 GP draft create/update: agent pin только `published`; gp-content/branching draft+published
- [x] 4.4 Integration tests: promote blocked/accepted scenarios

## 5. coin-ui — Platform entity pages

- [x] 5.1 Extract editors из `ComponentStudio.tsx` → reusable components (`GpContentEditor`, `BranchingModelEditor`)
- [x] 5.2 `PlatformBuildStackDetailPage` + edit route `/platform/build-stacks/:name/:version/edit`
- [x] 5.3 `PlatformBranchingModelDetailPage` + edit route `/platform/branching-models/:name/:version/edit`
- [x] 5.4 Catalog pages: create draft, publish, delete draft actions in-place
- [x] 5.5 Runtime catalog: убрать Studio references; published-only guidance

## 6. coin-ui — удалить Studio

- [x] 6.1 Удалить `ComponentStudio.tsx`, routes `/studio/*` из `App.tsx`
- [x] 6.2 Удалить «Open Studio» из `PlatformCatalogPage` и sidebar footer shortcut
- [x] 6.3 Обновить deep links в `GpReleaseDetail`, GP hub — Platform entity routes
- [x] 6.4 Обновить `coin-ui/README.md`

## 7. coin-ui — GP composition и promote gate

- [x] 7.1 `useGpCompositionEditor`: version pickers — draft+published для gp-content/branching; published only для agent
- [x] 7.2 Composition UI: status badges + draft pin warning text
- [x] 7.3 Promote button: disabled when draft pins; CTA links to Platform publish
- [x] 7.4 Canary line UI: warning при назначении GP draft с draft pins
- [x] 7.5 Surface API 409 blocking pins on promote failure

## 8. E2E и validation

- [x] 8.1 Обновить E2E scripts: убрать component `canary` asserts
- [x] 8.2 E2E: GP draft с draft gp-content pin → canary resolve green на pilot project
- [x] 8.3 E2E: GP promote blocked by draft pin → publish component → promote green
- [x] 8.4 `openspec validate platform-native-lifecycle --strict`
