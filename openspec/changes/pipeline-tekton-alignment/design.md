## Context

После `gp-release-embedded-pipeline` pipeline-inline **v3** хранится на GP release draft (`gp_release_pipeline_bodies`), composition — 2-pin (`agent` + `branching-model`). Модель v3 использует `pipeline.stages[]` с inline `run`/`build`/`publish` steps и **обязательным** `containerfile.body` на каждом buildkit step. Short-hash ids (`^[a-z0-9]{5,6}$`) затрудняют чтение manifest и соответствие Jenkins stage names.

Enabling team в explore-сессии согласовала **Tekton mental model** на Jenkins runtime:

| Tekton | Coin (authoring + runtime) |
|--------|---------------------------|
| Pipeline | GP release pipeline / Jenkins Pipeline |
| Task | Jenkins stage (`coinRunStage` → `coin-executor --task`) |
| Step | coin-executor action, Containerfile invoke, или `sh` |
| Workspace | Jenkins workspace на dynamic agent |
| PipelineRun | Jenkins build |
| TaskRun | не моделируется |
| Triggers | вне scope |

Связанные ADR: [pipeline-inline-build-stack.md](../../docs/adr/pipeline-inline-build-stack.md) (v3, будет supersede), [control-plane-v2.md](../../docs/adr/control-plane-v2.md).

## Goals / Non-Goals

**Goals:**

- Ввести **schemaVersion: 4** с `pipeline.tasks[]`, `runAfter`, step kinds (`coin`, `containerfile`, `sh`).
- **Containerfile catalog** на GP release body — named entries (`managed` body / `project` path), refs из steps.
- Semantic `task.id` (`validate`, `test`, `build`, `publish`) — совпадение с Jenkins stage и `coin-executor --task`.
- Materialize catalog + task graph в manifest; executor dispatch без inline duplicate containerfile bodies.
- UI: Composition → Pipeline (task graph) → Containerfiles catalog.
- v3 read-only compat + автоматическая миграция при save draft (pilot).

**Non-Goals:**

- Tekton Controller, CRD YAML, Triggers/EventListeners.
- Отдельный platform component pin для containerfile catalog.
- Изменение 2-pin composition или product config v2.
- Corp fleet migration prod GP.
- TaskRun / workspace PVC modeling.

## Decisions

### D1: schemaVersion 4 — `pipeline.tasks` вместо `pipeline.stages`

Author model:

```yaml
schemaVersion: 4
parameters: [...]
validateSchema: ...
containerfiles:
  - id: app
    kind: managed
    body: |
      FROM ...
  - id: liquibase
    kind: project
    path: docker/liquibase.Containerfile
pipeline:
  tasks:
    - id: validate
      name: Validate
      steps:
        - kind: coin
          action: validate
    - id: test
      name: Test
      runAfter: [validate]
      steps:
        - kind: containerfile
          ref: app
          run:
            output: test
    - id: build
      runAfter: [test]
      steps:
        - kind: coin
          action: build
          build:
            type: image
            containerfileRef: app
    - id: publish
      runAfter: [build]
      when: tag   # branching gate — unchanged semantics
      steps:
        - kind: coin
          action: publish
          publish:
            buildTaskId: build
```

**Альтернатива A:** оставить `stages` и только переименовать в UI.  
**Альтернатива B:** полный Tekton YAML import.  
**Выбор:** `tasks` + явный mapping в ADR; runtime остаётся Jenkins.

### D2: Containerfile catalog — секция GP release body

Catalog живёт в том же JSONB body, что и pipeline (не отдельная PG table на P0).

| `kind` | Смысл |
|--------|-------|
| `managed` | `body` — Coin-managed Containerfile (как сегодня inline) |
| `project` | `path` — BYO файл в product repo workspace |

Step `kind: containerfile` ссылается `ref: <catalog.id>` + `run` block (engine, output).

**Альтернатива:** отдельный platform component type `containerfile-bundle`.  
**Выбор:** catalog section на GP release — меньше pins, соответствует «embedded pipeline» decision.

### D3: Step kinds

| kind | P0 | Описание |
|------|-----|----------|
| `coin` | да | Бинарные примитивы executor: `validate`, `test`, `build`, `publish` |
| `containerfile` | да | Build/run через catalog ref |
| `sh` | pilot allowlist | Произвольный shell — только pre-approved actions в seed |

v3 `action: run|build|publish` **superseded** маппингом в `coin` + `containerfile`.

### D4: Semantic task ids

`task.id` SHALL match `^[a-z][a-z0-9-]{1,31}$` (semantic slug).  
`coin-executor` вызывается `run --task <id>` (alias `--stage` deprecated).

v3 short-hash stage ids: read-only; при save draft v3→v4 migration map по `stage.name` или positional default (`validate`, `test`, `build`, `publish`).

### D5: `runAfter` DAG

- P0: линейный default order + явный `runAfter[]`; coin-lib разворачивает в последовательные Jenkins stages.
- P1: `parallel` в coin-lib для tasks с одинаковым `runAfter` frontier (не в этом change).

Validation: DAG acyclic; все `runAfter` refs MUST exist.

### D6: Manifest materialization

Resolved manifest:

```yaml
containerfiles:
  - id: app
    contentRef: ...
    digest: sha256:...
pipeline:
  tasks: [...]  # steps materialized; containerfile steps carry contentRef inline
```

Nexus blob self-sufficient — без live PG для build path.

### D7: UI layout (GP release detail)

Порядок секций на draft:

1. Composition (2-pin)
2. Pipeline — task graph canvas + task cards (ordered steps)
3. Containerfiles — catalog editor
4. Parameters (collapsible или sidebar)

Preview panel — GP release pipeline preview API.

### D8: Migration v3 → v4

| Триггер | Поведение |
|---------|-----------|
| Read published v3 manifest | coin-api resolve: on-the-fly adapter (read compat) |
| Save draft | coin-api предлагает/применяет migration; `schemaVersion` → 4 |
| Seed scripts | пишут v4 shape напрямую |

Migration rules:
- `stages` → `tasks` (id from name or hash→semantic map)
- inline `containerfile.body` → dedupe into `containerfiles[]`, step → `kind: containerfile` + `ref`
- `publish.buildStepId` → `publish.buildTaskId`

Pilot: wipe/reseed допустим; fleet migration — post corp gate.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Breaking v3 drafts in flight | Migration on save + read compat for published |
| UI complexity (graph + catalog) | P0 linear graph; parallel P1 |
| `sh` step security | Allowlist in validate; no arbitrary shell in P0 prod GP |
| Semantic id collisions | Validate uniqueness; reserved ids documented |
| Manifest size (catalog vs inline) | Dedupe containerfiles; hash covers catalog |

## Migration Plan

1. JSON Schema `pipeline-inline.v4.schema.json` в coin-api + `internal/gpcontent/seed/`.
2. coin-api: validate/preview/migrate v3→v4; extend `gp_release_pipeline_bodies.schema_version`.
3. Manifest builder: materialize `containerfiles` + `pipeline.tasks`.
4. coin-executor: dispatch v4 step kinds; `--task` flag.
5. coin-lib: `coinPipeline` — topological sort `runAfter` → Jenkins stages.
6. coin-ui: refactor `GpReleasePipelineEditor` → task/step + catalog panels.
7. Seed `go-app`, `go-app-docker` stacks v4; `make seed-jenkins-lib`.
8. E2E `demo-go-app`, `demo-go-app-docker`.
9. ADR `docs/adr/pipeline-tekton-mapping.md`; archive sync specs.

Rollback: revert coin-api/UI; published v3 blobs остаются readable через adapter.

## Open Questions

| # | Вопрос | Статус | Варианты | Решение |
|---|--------|--------|----------|---------|
| Q1 | `sh` step в P0 seed stacks? | ⏳ | A: только `coin`+`containerfile` / B: `sh` для liquibase migrate | — |
| Q2 | `when` на task vs branching-only gate | ⏳ | A: task-level `when` / B: только manifest.branching | A (сохранить v3 semantics) |
| Q3 | UI graph library | ⏳ | A: custom React / B: React Flow | — |
