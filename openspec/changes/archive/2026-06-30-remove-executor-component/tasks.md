## 1. coin-api — registry & GP validate

- [x] 1.1 Удалить `executorPinForAgentStack`, `DerivedExecutorPin`, executor slot из `prepareGPRelease`
- [x] 1.2 Удалить `augmentCompositionWithDerivedExecutor` и `applyExecutorComposition`
- [x] 1.3 Убрать `executor` из `componentResolveMode`, `NextExecutorVersion`, legacy 4-slot composition
- [x] 1.4 Удалить `derivedExecutorPin` из GET component version response
- [x] 1.5 Migration `031_drop_executor_components.sql` — cleanup PG rows type=executor
- [x] 1.6 Обновить unit tests (`gp_release_prepare_test`, `agent_lifecycle_test`, `composition_loader_test`)

## 2. coin-api — manifest

- [x] 2.1 Убрать `executor` из `manifest/builder.go` и `Composition` struct
- [x] 2.2 Обновить `manifest.schema.json` — удалить секцию `executor` из required/properties
- [x] 2.3 Обновить OpenAPI examples (manifest, admin components)
- [x] 2.4 Обновить `builder_test.go`, resolve integration tests

## 3. coin-executor

- [x] 3.1 Удалить `Executor` из `internal/manifest/manifest.go` и test fixtures
- [x] 3.2 Удалить `bootstrap` subcommand (`cmd/coin-executor/main.go`, `internal/bootstrap/`)
- [x] 3.3 Обновить README / Jenkinsfile — без coin-api executor register

## 4. coin-lib

- [x] 4.1 Убрать `manifest.executor` из `coinLoadConfig.manifestToConfig`
- [x] 4.2 Обновить README (manifest fields list)

## 5. coin-ui

- [x] 5.1 Удалить derived executor section из `PlatformComponentReleaseDetail`
- [x] 5.2 Удалить `derivedExecutorPin` helper и API types
- [x] 5.3 Обновить `coin-ui-user-guide.md`

## 6. Scripts & E2E

- [x] 6.1 `publish-executor.sh` — Nexus upload only, без coin-api register (или deprecate с redirect в agent publish)
- [x] 6.2 `e2e-platform-component-hub.sh` — убрать assert `derivedExecutorPin`
- [x] 6.3 E2E: GP promote с `agent-30-06@V` без executor component (extend `e2e-gp-promote-gate.sh` или новый script)
- [x] 6.4 Seed/bootstrap scripts — не создавать executor component

## 7. Documentation & ADR

- [x] 7.1 Amend `docs/adr/coin-ci-runtime.md` §8 — agent-only, no manifest.executor
- [x] 7.2 Обновить `docs/agent-build-model.md`, `docs/config.md`, `docs/control-plane.md`, `docs/how-to/troubleshoot-ci.md`
- [x] 7.3 Supersede `docs/adr/gp-composition-four-components.md` reference to executor slot (if still linked)

## 8. Validation

- [x] 8.1 `openspec validate remove-executor-component`
- [x] 8.2 `go test ./...` в coin-api, coin-executor
- [x] 8.3 Manual: GP draft `agent-30-06@1.2.0` + publish gc/bm → promote succeeds
