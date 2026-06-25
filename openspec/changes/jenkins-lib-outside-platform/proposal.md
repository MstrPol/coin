## Why

Платформа Coin управляет GP composition и manifest для **coin-executor**, но не должна управлять Jenkins Shared Library (`coin-lib`). Lib — внешний CI glue: версия задаётся в Jenkins org/CASC (`@Library('coin-lib@…') _`), а не через operator UI или `platform_settings`. Текущая модель дублирует ответственность (platform lib pin, `/platform/jenkins-lib`, секция `lib` в manifest, component type `lib` в registry) и противоречит ADR [jenkins-lib-http-nexus.md](../../docs/adr/jenkins-lib-http-nexus.md), где product Jenkinsfile не использует platform API для выбора lib на build path. Полный вынос lib из coin-api упрощает ментальную модель Platform и готовит exit path от Jenkins без переделки control plane.

## What Changes

- **BREAKING:** Удалить component type `lib` из coin-api registry (таблицы, admin API, Studio routes, `next-version`).
- **BREAKING:** Удалить `platform_settings.runtime` / `PlatformRuntime` (lib pin) из API и UI.
- **BREAKING:** Resolve manifest **не** содержит секцию `lib`; `augmentCompositionWithPlatformRuntime` для lib убирается.
- **BREAKING:** Удалить `GET /v1/golden-paths/{name}/version` (LibraryVersion bootstrap API).
- UI: убрать `/platform/jenkins-lib`, lib pin из Platform settings и lib из каталога `/platform/runtime` (остаются только `agent` + `executor`).
- **coin-lib:** publish только в Nexus (ZIP + HTTP retriever); регистрация через coin-api admin API убирается.
- **ADR:** новый `docs/adr/jenkins-lib-outside-platform.md`; amend cross-ref в `jenkins-lib-http-nexus.md` и `gp-component-package-model.md`.
- Spec `platform-runtime-line` — **REMOVED** (capability больше не актуален).
- Bootstrap/seed: `seed-jenkins-lib` публикует lib в Nexus без PG registry; E2E assertions на `composition.type==lib` убираются.

## Non-Goals

- Смена product Jenkinsfile контракта (`@Library` + `coinPipeline()`).
- Перенос build logic в coin-lib (остаётся coin-executor only).
- Удаление репозитория `coin-lib/` или Jenkins HTTP Shared Library retriever.
- Corp gate / fleet rollout.
- Объединение `agent`+`executor` в один component type.
- GitHub Actions / Tekton glue (только подготовка границы).

## Capabilities

### New Capabilities

- `jenkins-lib-boundary`: граница ответственности — lib вне coin-api/control plane; Nexus-only publish; Jenkins SoT для версии glue.

### Modified Capabilities

- `platform-runtime-catalog`: `/platform/runtime` — только agent и executor; без lib pin banner и без type `lib` в фильтре.
- `platform-build-stacks`: без изменений по gp-content (подтверждение: lib не в Platform IA).
- `gp-composition-two-slot`: resolve materialization без platform lib injection; manifest без `lib`.
- `gp-publish-flows`: убрать валидацию platform lib pin при draft create.
- `ui-enabling-shell`: nav Platform без Jenkins library; Platform settings без runtime section.

### Removed Capabilities

- `platform-runtime-line`: platform-level lib pin и resolve injection — superseded этим change.

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-api** | migrations (drop lib components, drop runtime column), store/resolve/manifest, OpenAPI, admin routes |
| **coin-ui** | nav, PlatformRuntimePage, PlatformSettings, удалить PlatformJenkinsLibPage, PublishWizard runtime validation |
| **coin-lib** | `publish-lib.sh`, `Jenkinsfile` — Nexus-only, без coin-api register |
| **docker/scripts** | `seed-jenkins-lib-stack.sh`, e2e scripts |
| **docs/adr** | новый ADR + README index |
| **openspec/specs** | remove `platform-runtime-line`; deltas для modified capabilities |

См. [jenkins-lib-http-nexus.md](../../docs/adr/jenkins-lib-http-nexus.md), [gp-component-package-model.md](../../docs/adr/gp-component-package-model.md), [coin-lib-scope.mdc](../../.cursor/rules/coin-lib-scope.mdc).
