# ADR: Jenkins Shared Library — coin-lib

**Статус:** accepted; **amended** — gp-content package path removed (embedded GP pipeline)  
**Дата:** 2026-06  

> **Live:** `coin-lib` = Jenkins glue only. GP pipeline content — embedded на GP release + `coin-api` seed.  
> Исторический заголовок «coin-lib + gp-content» и 5-slot composition ниже — superseded. См. [gp-embedded-pipeline](gp-embedded-pipeline.md), [jenkins-lib-outside-platform](jenkins-lib-outside-platform.md).

## Контекст

Product Jenkinsfile вырос до ~100 строк: resolve manifest, fallback Nexus, download pipeline-bundle ZIP, verify SHA, unzip, load Groovy kernel, credentials, pod, stages. Jenkins-specific glue смешался с bootstrap и дублирует возможности Jenkins Shared Library.

Параллельно pipeline-bundle объединял две разные ответственности:
- **Jenkins glue** — pod template, credentials binding, stage orchestration
- **GP content** — scripts, Dockerfile, validate schema

## Решение

### GP composition (5 slots)

```
jnlp + agent + executor + lib/coin-lib + gp-content/{golden-path}
```

Hard cut — `pipeline-bundle` не поддерживается.

### Product Jenkinsfile

```groovy
@Library('coin-lib@1.0.0') _
coinPipeline()
```

### Разделение репозиториев

| Репозиторий | Роль |
|-------------|------|
| `coin-lib/` | Jenkins Shared Library — только glue (resolve, pod, creds, вызов coin-executor) |
| GP pipeline | Embedded на GP release; bootstrap seed в `coin-api/internal/gpcontent/seed/` |

> **Superseded:** отдельная папка/repo `coin-gp-content/` (ZIP → Nexus gp-content packages).

### Версионирование lib

- **Phase 1 (local pilot):** `coin-lib` из Gitea `coin/coin-lib` tag `1.0.0` через Modern SCM retriever
- **Target (prod):** immutable ZIP из Nexus через HTTP Shared Libraries retriever

### Platform API

> **Superseded (2026-06):** lib registry, platform lib pin, manifest `lib` section и `LibraryVersion` API удалены — см. [jenkins-lib-outside-platform.md](jenkins-lib-outside-platform.md). Product Jenkinsfile по-прежнему не использует coin-api для выбора lib на build path.

- Manifest **не** содержит `orchestration.bundle`; scripts/schema refs из `gp-content`
- `coin-lib` исполняет Jenkins stages динамически из `manifest.pipeline.stages`

### Deliverables (product contract)

> **Superseded (2026-07):** product-level `deliverables` удалены из `.coin/config.yaml`. GP/Build Stack полностью задаёт P0 deliverables (`image`, `liquibase-image`, `artifact`), а product config хранит только GP pin, project identity и Jenkins glue.

### Extension policy

- P0: declarative `artifact.sources` (explicit paths only)
- P1 hooks: docs only
- Запрет custom Jenkins actions / stage overrides

## Последствия

- `coin-lib-scope` получает исключение: Shared Library разрешена **только** как Jenkins glue
- `coin-pipeline-bundles/` удаляется; Groovy — в `coin-lib/`; content → embedded GP pipeline + coin-api seed (исторически через `coin-gp-content/`, папка удалена)
- ADR `gp-pipeline-bundle-layer.md` и `gp-composition-four-components.md` superseded этим ADR
- UI не показывает raw gp-content artifacts на GP release detail

## Отклонённые альтернативы

| Альтернатива | Почему отклонена |
|--------------|------------------|
| Оставить fat Jenkinsfile + pipeline-bundle ZIP | Дублирование, сложный bootstrap, смешение glue и content |
| Dynamic `httpRequest` bootstrap в product repo | Возвращает platform API URL, credentials и JSON parsing в каждый product repo |
| Backwards compatibility с pipeline-bundle | Hard cut для local pilot; один активный контракт |
| Runtime reload `coin-lib` после manifest resolve | Ненадёжный контракт Jenkins Pipeline: та же Shared Library уже загружена |
| Shared Library с бизнес-логикой сборки | Нарушает coin-lib-scope; логика в coin-executor |
