## 1. ADR и документация (blocking)

- [x] 1.1 Создать `docs/adr/jenkins-lib-outside-platform.md` (граница, SoT lib vs manifest, последствия)
- [x] 1.2 Обновить `docs/adr/README.md` — индекс нового ADR
- [x] 1.3 Amend `docs/adr/jenkins-lib-http-nexus.md` — platform API / lib registry superseded
- [x] 1.4 Amend `docs/adr/gp-component-package-model.md` — убрать `lib` из platform component types

## 2. coin-api — migration

- [x] 2.1 Migration: удалить `component_*` rows для `type='lib'`; удалить `gp_composition` rows `component_type='lib'`
- [x] 2.2 Migration: drop `platform_settings.runtime` column (или вся таблица runtime jsonb)
- [x] 2.3 Проверить seed/bootstrap на чистой БД после migration

## 3. coin-api — remove lib from control plane

- [x] 3.1 Удалить `PlatformRuntime`, `GetPlatformRuntime`, lib injection из `store.go` / `gp_release_prepare.go`
- [x] 3.2 Упростить resolve: executor-from-agent only; убрать `lib` из `manifest/builder.go` и `composition_loader.go`
- [x] 3.3 Удалить `LibVersionFromComposition`, `NextLibVersion`, admin routes `components/lib/*`
- [x] 3.4 Удалить `GET /v1/golden-paths/{name}/version` (LibraryVersion)
- [x] 3.5 OpenAPI: убрать `PlatformRuntime`, lib schemas, LibraryVersion path; `manifest.schema.json` без `lib`
- [x] 3.6 `PlatformSettings` API — только Nexus fields
- [x] 3.7 Unit/integration tests: resolve без `lib`; draft create без platform runtime

## 4. coin-ui — Platform IA cleanup

- [x] 4.1 Удалить `PlatformJenkinsLibPage`, route, nav item «Jenkins library»
- [x] 4.2 Redirect `/platform/jenkins-lib` → `/platform/runtime`
- [x] 4.3 `PlatformRuntimePage` — убрать lib pin banner; catalog types agent+executor only
- [x] 4.4 `PlatformSettings` — убрать runtime/lib section; API types без `PlatformRuntime`
- [x] 4.5 `PublishWizard` — убрать `validateRuntimePins` и lib warnings
- [x] 4.6 Обновить `coin-ui/README.md`, `docs/coin-ui-user-guide.md`
- [x] 4.7 Убрать stale copy (CreateGPProfile, GpReleaseDetail) про platform lib / runtime

## 5. coin-lib — Nexus-only publish

- [x] 5.1 `publish-lib.sh` — убрать coin-api register; только Nexus upload
- [x] 5.2 `coin-lib/Jenkinsfile` — убрать `nextLibVersion` coin-api call; semver via BUMP param
- [x] 5.3 Обновить `coin-lib/README.md` — SoT версии, без registry

## 6. docker seed & E2E

- [x] 6.1 `seed-jenkins-lib-stack.sh` — убрать PUT platform settings runtime; lib publish Nexus-only
- [x] 6.2 E2E scripts: убрать asserts `composition.type==lib` и manifest `.lib` (`e2e-jenkins-lib.sh`, `e2e-mvp1.sh`, `e2e-branching-canary-resolve.sh`)
- [ ] 6.3 `make seed-jenkins-lib` + `make e2e-demo-go-app` green
- [x] 6.4 `openspec validate jenkins-lib-outside-platform --strict`

## 7. Docs cross-ref

- [x] 7.1 `docs/control-plane.md`, `docs/golden-paths.md` — lib вне platform scope
- [x] 7.2 `docs/how-to/publish-gp-release.md`, runbooks — без lib registry steps
