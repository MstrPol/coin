## Context

**Текущее состояние (после `pipeline-inline-steps`):**

- Pipeline-inline v3 живёт в `gp-content` component package (`content.yaml` structured model).
- GP release composition — 3 pin: `agent`, `gp-content`, `branching-model`.
- Authoring: Platform → Build stacks → edit → register → promote component → pin в GP draft → promote GP.
- `gp_profiles.name` может отличаться от `gpContentName` (seed: `gp-01-07` → `go-app`).
- Product resolve не меняется: manifest собирается из composition + materializer.

**Зафиксированные product decisions (platform lead, explore 2026-07-03):**

| Решение | Значение |
|---------|----------|
| Pipeline reuse между GP profiles | **Нет** |
| Pipeline semver отдельно от GP | **Нет** — меняется только с bump GP version |
| Author | **Enabling team** в GP hub |
| Branching | **Отдельный** platform component |
| gp-content hard cut | **Да** — проект в концепте, fleet migration не нужен |

Связанные ADR: [pipeline-inline-build-stack.md](../../docs/adr/pipeline-inline-build-stack.md) (формат v3), [gp-component-package-model.md](../../docs/adr/gp-component-package-model.md) (supersede gp-content section), [control-plane-v2.md](../../docs/adr/control-plane-v2.md).

## Goals / Non-Goals

**Goals:**

- GP release = единственная primary entity для build contract (parameters + pipeline).
- Composition = 2 внешних pin + embedded pipeline.
- Один promote workflow на GP release.
- Authoring UX: GP hub → release detail → Pipeline tab.
- Hard cut: нет `gp-content` в registry, UI, seed aliases.

**Non-goals:**

- Shared pipeline templates между GP profiles.
- Отдельный semver для pipeline payload.
- Platform → Build stacks catalog (даже read-only).
- Migration legacy gp-content rows в PG (pilot wipe/reseed допустим).
- Изменение manifest shape для coin-executor (sections те же).

## Decisions

### D1: Pipeline storage — `gp_release_pipeline_bodies`

Новая таблица (или расширение `gp_artifact_bodies`) для structured pipeline-inline model на GP release draft:

```
gp_release_pipeline_bodies (
  gp_release_id  FK → gp_releases,
  schema_version INT NOT NULL DEFAULT 3,
  body           JSONB NOT NULL,   -- parameters, validateSchema, pipeline
  sha256         TEXT NOT NULL,
  updated_at     TIMESTAMPTZ
)
```

**Published SoT:** pipeline materialized **внутри** Nexus manifest blob при GP promote (как сейчас subset из gp-content manifest). PG draft body — edit SoT до promote.

**Альтернатива:** переиспользовать `gp_artifact_bodies` с key `pipeline`.  
**Выбор:** dedicated table или single key `pipeline` в `gp_artifact_bodies` — implementation detail; контракт: one pipeline body per GP release version.

### D2: Composition — 2-pin only

`gp_composition` хранит только:

| Slot | Type |
|------|------|
| `agent` | `agent` |
| `branching-model` | `branching-model` |

Draft API: `agentStackName`, `branchingModelName`, `composition: { agent, branching-model }`.  
Поле `gpContentName` **удаляется**.

Resolve: `loadComposition` + `loadPipelineBody(release)` → manifest builder.

### D3: GP profile name = pipeline identity

`gp_profiles.name` — единственный идентификатор pipeline family. Запрещён сценарий «profile `xxx` pins `go-app` content».  
`go-app` и `go-app-docker` — **разные GP profiles** с разными embedded pipelines.

### D4: Preview API scope

```
POST /v1/admin/golden-paths/{name}/versions/{version}/pipeline/preview
```

Request: structured pipeline-inline v3 model (или read from draft body).  
Response: resolved manifest subset + validation issues (как сегодня gp-content preview).

Удалить: `POST /v1/admin/gp-content/preview`.

### D5: UI — Pipeline tab on release detail

- Перенести `GpContentEditor` → `GpReleasePipelineEditor` на `/gp/{name}/releases/{version}` (draft only).
- Удалить family `build-stacks` из `platformFamilyConfig`, routes, nav.
- `GpCompositionForm`: 2 picker'а (agent, branching).
- Promote gate: agent + branching published + pipeline valid (без gp-content pin check).

### D6: coin-gp-content repo role

Остаётся **git seed source** для bootstrap (`stacks/go-app/`, `stacks/go-app-docker/`).  
`publish-content.sh` → deprecate; seed script читает stack YAML и пишет в `gp_release_pipeline_bodies` при создании GP release.

### D7: Nexus layout

Не создавать `coin/gp-content/{name}/{version}/` packages.  
Pipeline digest входит в GP manifest `manifestHash` при promote.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Большой blast radius в coin-api/UI | Hard cut на pilot; focused tests + reseed |
| Путаница с archived specs (gp-composition-two-slot) | Новый capability `gp-release-two-pin`; archive sync помечает старый spec superseded |
| `gp_artifact_bodies` legacy dual-write runbook | Обновить runbook: pipeline только на release |
| Имя `build-stack-pipeline-ui` spec устарело | Delta переименует scope на GP release (без rename capability id на pilot) |

## Migration Plan

1. DB migration: drop gp-content composition rows; add pipeline bodies table; purge `component_versions` where type=`gp-content`.
2. coin-api: new endpoints + resolve path; remove gp-content component handlers.
3. coin-ui: pipeline on release detail; remove build-stacks routes.
4. Seed: `make seed-jenkins-lib` создаёт GP profiles с embedded pipeline.
5. Docs + ADR update; `openspec validate --strict`.
6. E2E: `demo-go-app`, `demo-go-app-docker` green.

**Rollback:** pilot-only — `make wipe-gp` + reseed prior git tag if needed.

## Open Questions

| # | Вопрос | Статус | Варианты | Решение |
|---|--------|--------|----------|---------|
| Q1 | Dedicated table vs `gp_artifact_bodies` key | ⏳ | A: `gp_release_pipeline_bodies` / B: artifact key `pipeline` | — (implementation) |
| Q2 | Rename spec `build-stack-pipeline-ui` → `gp-release-pipeline-ui` | ⏳ | A: rename at archive / B: keep id, change requirements | — |
