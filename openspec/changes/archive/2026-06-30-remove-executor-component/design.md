## Context

Текущий поток:

```
GP agent pin @V  →  validate: executor/coin-executor@V in PG  →  resolve: manifest.executor{version,url,sha256}
Jenkins pod       →  runtime.image + digest only; coin-executor binary baked in image
```

Ошибка из explore: `agent-30-06@1.2.0` published, `executor/coin-executor@1.2.0` отсутствует → GP promote blocked.

ADR [coin-ci-runtime](../../docs/adr/coin-ci-runtime.md) уже фиксирует baked binary и three-pin composition, но код и спеки сохранили legacy derive из change `runtime-agent-registry`.

**Решение (вариант B):** agent — единственный runtime component; `manifest.executor` удаляется полностью.

## Goals / Non-Goals

**Goals:**

- GP create/save/promote/resolver не обращаются к `executor` component.
- Resolved manifest v1: `runtime` из agent metadata; без секции `executor`.
- UI/API не показывают derived executor pin.
- E2E: custom agent profile promote GP без отдельного executor publish.
- Документация и ADR согласованы.

**Non-Goals:**

- Менять semver agent versioning или promote gate (image+digest).
- Удалять Nexus path для executor binary (может использоваться при сборке agent image в CI).
- Dual path / fallback на старый manifest с `executor`.

## Decisions

### D1. Agent = полный CI runtime stack (единственный pin)

GP composition slot `agent` покрывает образ + baked `coin-executor`. Отдельный registry type `executor` **не существует**.

### D2. manifest.executor — удалить (вариант B)

```json
// Было (required)
"executor": { "version", "url", "sha256" }

// Станет
// (секции нет)
```

`manifestVersion` остаётся `1` (hard cut local pilot, без v2 bump).

**Альтернатива A (version-only)** — отклонена platform lead.

### D3. coin-api: убрать derive pipeline

| Функция | Действие |
|---------|----------|
| `executorPinForAgentStack` | Удалить |
| `DerivedExecutorPin` | Удалить |
| `prepareGPRelease` executor slot merge | Удалить |
| `augmentCompositionWithDerivedExecutor` | Удалить |
| `applyExecutorComposition` | Удалить |
| `legacyFullComposition` (4-slot) | Удалить или reject |
| `componentResolveMode` для executor | Удалить |
| `NextExecutorVersion` | Удалить |

GP validate: только `gpSlots` (3 pins).

### D4. manifest.builder

`Build()` не добавляет `executor`. `manifest.Composition` struct — убрать `ExecutorVersion`, `ExecutorURL`, `ExecutorSHA256`.

### D5. coin-executor CLI

- `internal/manifest.Executor` struct — удалить поле из `Manifest`.
- `bootstrap download` — удалить subcommand (superseded).
- Validate/load tests — fixtures без `executor`.

### D6. coin-lib

`coinLoadConfig.manifestToConfig` — убрать блок `manifest.executor?.url`.

### D7. publish-executor.sh

Прекратить `POST /v1/admin/components/executor/...`. Скрипт может остаться как **Nexus upload only** для Jenkins agent image build (вне scope registry) или merge в agent publish pipeline — зафиксировать в README.

### D8. UI

Убрать секцию «Derived executor pin» с `PlatformComponentReleaseDetail`. Удалить `derivedExecutorPin` из `platformComponentPaths.ts` и API response field.

### D9. Data migration (local pilot)

Migration `028_drop_executor_components.sql`:

- DELETE `component_versions` / `components` WHERE `type = 'executor'`.
- Не трогать `gp_composition` (executor не хранится в 3-pin model).

Rollback: re-seed executor@1.0.0 из bootstrap script.

### D10. OpenSpec supersede

Requirements про executor derive в `runtime-agent-registry`, `platform-component-lifecycle`, `gp-composition-two-slot`, `platform-component-hub`, `runtime-documentation` — REMOVED/MODIFIED в delta specs.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| External consumer читает `manifest.executor` | Local pilot hard cut; grep repo + docs |
| Legacy manifests в Nexus cache | Re-resolve GP releases после deploy |
| Jenkinsfile coin-executor всё ещё register executor | Обновить Jenkinsfile + publish script в том же change |
| Audit «какая версия executor в CI» | Agent pin version = image tag; `coin-executor version` в pod |

## Migration Plan

1. coin-api + migration (deploy first).
2. Re-publish или re-resolve активные GP drafts.
3. coin-executor + coin-lib (agent pod).
4. coin-ui.
5. E2E: `e2e-platform-component-hub.sh`, `e2e-gp-promote-gate.sh`, GP draft с custom agent.

Rollback: revert deploy; re-seed executor component для старого resolve (не рекомендуется — держать forward).

## Open Questions

| # | Вопрос | Статус | Решение |
|---|--------|--------|---------|
| Q1 | manifest.executor: version-only vs remove | ✅ | **B — remove** (platform lead) |
| Q2 | Оставить publish-executor.sh для Nexus only? | ⏳ | Рекомендация: да, без coin-api register; уточнить в tasks |
| Q3 | manifestVersion bump to 2? | ✅ | Нет — hard cut v1 без executor section |
