# ADR: Jenkins Shared Library вне scope control plane

**Статус:** accepted  
**Дата:** 2026-06  
**Supersedes (частично):** [jenkins-lib-http-nexus.md](jenkins-lib-http-nexus.md) (Platform API / lib registry), `platform-runtime-line` OpenSpec capability  
**Связанные ADR:** [control-plane-v2.md](control-plane-v2.md), [build-engine-contract.md](build-engine-contract.md), [gp-component-package-model.md](gp-component-package-model.md)

## Контекст

Coin control plane (coin-api + coin-ui) управляет GP composition и manifest для **coin-executor**: agent stack, gp-content, branching-model. Jenkins Shared Library (`coin-lib`) — CI glue: resolve manifest, K8s pod, credentials, orchestration stages. Версия lib задаётся в Jenkins org (`@Library('coin-lib@1.0.0') _`), не через operator UI.

После change `gp-profile-metadata-model` в платформе остались артефакты управления lib: component type `lib` в registry, `platform_settings.runtime.lib`, секция `lib` в manifest, `/platform/jenkins-lib`, `LibraryVersion` API. Это дублирует Jenkins SoT и противоречит границе «lib развивается отдельно от платформы».

## Решение

### Граница ответственности

```
┌──────────────────────┐         ┌──────────────────────┐
│  Jenkins org / CASC  │         │  Coin control plane │
│  @Library version    │         │  GP draft 3-pin     │
│  HTTP retriever      │         │  resolve → manifest │
│  Nexus lib ZIP       │         │  (executor path)    │
└──────────┬───────────┘         └──────────┬──────────┘
           │                                │
           │  coinPipeline()                │  GET .../manifest
           └────────────┬───────────────────┘
                        ▼
                 coin-executor stages
```

| Concern | SoT | Не в scope платформы |
|---------|-----|----------------------|
| lib version | Jenkins Shared Library config + Nexus immutable ZIP | coin-api registry, platform settings, operator UI |
| GP build manifest | coin-api resolve | секция `lib` в manifest |
| gp-content, agent, branching | coin-api component registry + GP composition | — |

### coin-api

- Удалить component type `lib` из PostgreSQL registry и admin API.
- Удалить `platform_settings.runtime` (lib pin).
- Resolve manifest **без** top-level `lib`; materialize executor from agent stack only.
- Удалить `GET /v1/golden-paths/{name}/version` (LibraryVersion).

### coin-ui

- Удалить `/platform/jenkins-lib` и lib pin из Platform settings.
- `/platform/runtime` — каталог **agent + executor** only.

### coin-lib

- Publish path: **Nexus ZIP only** (`publish-lib.sh`); без POST register в coin-api.
- CI semver: параметр версии / local bump script; без `next-version` API.

### Product contract (без изменений)

```groovy
@Library('coin-lib@1.0.0') _
coinPipeline()
```

## Последствия

- Enabling team не управляет lib через Platform; runbook для lib — Jenkins admin + Nexus.
- Audit lib version per GP release **не** в manifest; fleet audit через Jenkins job metadata.
- `gp-component-package-model`: platform component types = `agent`, `executor`, `gp-content`, `branching-model` (без `lib`).
- OpenSpec `platform-runtime-line` removed при archive change `jenkins-lib-outside-platform`.
- Local bootstrap: `make seed-jenkins-lib` публикует lib в Nexus, не в PG.

## Отклонённые альтернативы

| Альтернатива | Почему отклонена |
|--------------|------------------|
| UI-only removal (lib остаётся в registry/API) | Ложная модель: оператор видит lib в API, но не в UI |
| Platform lib pin без registry | Дублирование Jenkins SoT без полной пользы |
| lib в manifest для audit | coin-lib не читает; смешивает CI bootstrap с executor contract |
| Deprecate 410 вместо hard cut | Local pilot — один активный контракт |
