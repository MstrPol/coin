## Context

После change `gp-profile-metadata-model` platform runtime сузился до **lib-only pin** в `platform_settings.runtime`, а `/platform/runtime` показывал lib pin + каталог agent/executor/lib. Это противоречит принятой границе: **coin-lib — Jenkins glue вне control plane**; версия lib задаётся в Jenkins (`@Library('coin-lib@1.0.0') _`), manifest нужен executor'у (runtime, pipeline, branching), а не bootstrap'у Shared Library.

Текущие точки coupling:

| Слой | lib coupling |
|------|----------------|
| PG `components` type=`lib` | registry + artifact bodies |
| `platform_settings.runtime.lib` | operator pin |
| resolve `augmentCompositionWithPlatformRuntime` | inject lib в manifest |
| `manifest.schema.json` `lib` section | contract |
| `GET /v1/golden-paths/{name}/version` | LibraryVersion API |
| coin-ui `/platform/jenkins-lib`, settings, runtime banner | operator surface |
| `coin-lib/scripts/publish-lib.sh` | POST register в coin-api |

Stakeholders: enabling team (Platform IA), platform lead (ADR), local pilot E2E.

## Goals / Non-Goals

**Goals:**

- Полностью убрать `lib` из coin-api metadata model и operator UI.
- Manifest resolve = GP 3-pin + derived executor only (без `lib` section).
- coin-lib publish path = Nexus ZIP only; Jenkins HTTP retriever — единственный consumer версии.
- Новый ADR фиксирует границу; supersede части `platform-runtime-line` spec.
- `/platform/runtime` остаётся каталогом **agent + executor** без lib.

**Non-Goals:**

- Изменение product Jenkinsfile (`@Library` + `coinPipeline()`).
- Удаление репозитория `coin-lib/`.
- Corp fleet rollout.
- Tekton/GitHub Actions glue implementation.

## Decisions

### D1. lib не является platform component

**Решение:** удалить type `lib` из component registry (components, component_versions, bodies, admin API, OpenAPI schemas).

**Альтернатива:** оставить registry для audit — отклонена: создаёт ложное впечатление, что enabling team управляет lib через Platform.

**Миграция:** migration SQL удаляет rows `type='lib'`; legacy `platform-starter` cleanup уже частично в seed.

### D2. Удалить platform_settings.runtime

**Решение:** `PlatformSettings` = Nexus fields only (`nexusMavenBase`, `nexusCredentialsId`). Колонка `runtime` jsonb drop или оставить null deprecated — prefer **drop** в одной migration.

**Альтернатива:** пустой `{}` runtime — отклонена: мёртвое поле в API.

### D3. Manifest без секции lib

**Решение:** `manifest.builder` не emit `lib`; `composition_loader` убрать `applyLibComposition`; `augmentCompositionWithPlatformRuntime` переименовать/упростить до executor-from-agent only.

**Обоснование:** coin-lib не читает свою версию из manifest; ADR jenkins-lib-http-nexus: build path не использует LibraryVersion API.

### D4. Удалить LibraryVersion API

**Решение:** удалить `GET /v1/golden-paths/{name}/version` и `resolve.LibraryVersion`.

**Альтернатива:** deprecate 410 — отклонена при local pilot hard cut.

### D5. coin-lib Nexus-only publish

**Решение:** `publish-lib.sh` — только ZIP upload в Nexus (`maven-releases` path per `lib/maven-url.sh`); убрать POST `/v1/admin/components/lib/...`. `coin-lib/Jenkinsfile` — убрать `nextLibVersion` coin-api call; semver bump локально или через git tag + script param (как сейчас BUMP param, но без API alloc).

**Версионирование lib SoT:** git tag + Nexus immutable artifact + Jenkins CASC HTTP retriever URL с версией.

### D6. Platform UI cleanup

**Решение:**

- Удалить route `/platform/jenkins-lib`, nav item, `PlatformJenkinsLibPage.tsx`.
- `PlatformRuntimePage` — убрать lib pin banner; catalog types `["agent", "executor"]` only.
- `PlatformSettings` — убрать runtime section; только Nexus.
- `PublishWizard` — убрать `validateRuntimePins` и lib warnings.
- Redirect `/platform/jenkins-lib` → `/platform/runtime` (optional, 1 release) или 404 — prefer **redirect** для bookmarks.

### D7. ADR jenkins-lib-outside-platform

**Решение:** новый ADR в `docs/adr/jenkins-lib-outside-platform.md`:

1. Контекст — lib вне control plane scope
2. Граница ответственности (diagram)
3. SoT для lib version (Jenkins org)
4. SoT для build manifest (coin-api resolve)
5. Последствия для gp-component-package-model (lib не в UI-first component types)
6. Amend note в `jenkins-lib-http-nexus.md` (platform API section superseded)

### D8. Legacy 5-slot composition

**Решение:** `isLegacyFullComposition` / `legacyFullCompositionSlots` — убрать `lib` key из legacy path или оставить read-only для старых rows в DB без re-resolve. Prefer: **migration** очистить lib rows из `gp_composition` where `component_type='lib'`; validation rejects lib in new drafts (уже есть).

### D9. E2E / seed

**Решение:** `seed-jenkins-lib-stack.sh`:

- Вызывает `publish-lib.sh` (Nexus only)
- Убирает PUT platform settings runtime
- E2E scripts: убрать jq asserts на `composition.type==lib` и manifest `.lib`

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Jenkins CASC не указывает ту же версию lib, что в Nexus | Runbook: seed + CASC pin same version; E2E checks Jenkins build, not manifest.lib |
| Потеря audit trail lib version per GP release | Accept: lib version не привязана к GP; fleet audit через Jenkins job metadata |
| Breaking external consumers of LibraryVersion API | Hard cut local pilot; document removal in OpenAPI changelog |
| coin-lib CI без next-version API | Semver via Jenkins params / git tags; document in coin-lib README |

## Migration Plan

1. **ADR** — merge first (decision record).
2. **coin-api migration** — delete lib components; drop runtime column; clean gp_composition lib rows.
3. **coin-api code** — resolve, manifest, admin routes, platform settings.
4. **coin-ui** — remove pages/fields.
5. **coin-lib** — publish script + Jenkinsfile.
6. **docker seed + e2e** — adjust scripts.
7. **openspec** — archive change syncs spec deltas; remove `platform-runtime-line` from main specs.
8. **Verify** — `make seed-jenkins-lib` + `make e2e-demo-go-app` green.

**Rollback:** restore migration + redeploy previous images (local pilot only).

## Open Questions

| # | Вопрос | Статус | Варианты | Решение |
|---|--------|--------|----------|---------|
| Q1 | Redirect `/platform/jenkins-lib`? | ✅ | 404 / redirect to `/platform/runtime` | Redirect (bookmarks) |
| Q2 | Drop `runtime` column vs null | ✅ | drop / keep null | Drop column |
| Q3 | coin-lib semver without next-version API | ✅ | git tag only / manual BUMP param | Keep Jenkins BUMP param + local calc |
