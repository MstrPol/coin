## Why

После `gp-release-embedded-pipeline` pipeline-inline v3 живёт на GP release, но модель остаётся фрагментированной: `stages` без явной Tekton-онтологии, Containerfile дублируется inline в каждом step, нет каталога нескольких Containerfile на GP, смешаны доменные `run/build/publish` и границы Jenkins stage. Enabling team и визуальный редактор нуждаются в **стройной иерархии Pipeline → Task → Step**, согласованной с Jenkins runtime (не с Tekton Controller).

## What Changes

- **BREAKING:** `schemaVersion: 4` — pipeline-inline модель с `pipeline.tasks[]` (бывш. `stages`), явными `steps[]` и каталогом `containerfiles`.
- **BREAKING:** Containerfile **не inline** в step; named catalog на GP release (`managed` body или `project` path), step ссылается по `ref`.
- **BREAKING:** Step kinds: `coin` (бинарные примитивы executor), `containerfile` (managed/project invoke), `sh` (произвольный shell — pilot allowlist).
- Tekton entity mapping (authoring/runtime): Pipeline = GP/Jenkins pipeline; Task = Jenkins stage; Step = executor action; Workspace = agent checkout; PipelineRun = Jenkins build.
- `runAfter` на Task для DAG; P0 — линейный порядок + явные зависимости; P1 — Jenkins `parallel` в coin-lib.
- Semantic `task.id` (например `validate`, `test`, `build`) для `coin-executor run --task`; v3 hash ids — read-only legacy.
- Supersede v3 requirement «containerfile only inline in step».
- UI: Composition → Pipeline (task graph) → Containerfiles catalog; task card для ordered steps.
- coin-api: validate/preview/materialize v4; migration reader v3 → v4.
- coin-executor: dispatch `coin` / `containerfile` / `sh` steps внутри task.
- Triggers / Tekton TaskRun / PipelineRun CRDs — **вне scope**.

## Capabilities

### New Capabilities

- `pipeline-tekton-model`: schema v4, Tekton→Coin mapping, `pipeline.tasks`, `runAfter`, step kinds (`coin`, `containerfile`, `sh`).
- `gp-containerfile-catalog`: named containerfiles на GP release (`kind: managed|project`), materialize в manifest, refs из steps.

### Modified Capabilities

- `pipeline-inline-model`: supersede v3 inline-only; делегировать v4 в `pipeline-tekton-model`.
- `gp-embedded-pipeline`: GP release body включает containerfile catalog + pipeline tasks.
- `manifest-pipeline-inline`: materialize catalog contentRefs и task graph для executor/Jenkins.
- `build-stack-pipeline-ui`: редактор tasks/steps, catalog containerfiles, pipeline canvas.
- `build-engine`: step dispatch v4 (`coin`, `containerfile`, `sh`) и `--task` mapping.
- `gp-entity-hub`: layout GP release detail (composition, pipeline, containerfiles).
- `jenkins-lib-boundary`: coinPipeline разворачивает task DAG в Jenkins stages (linear P0, parallel P1).

## Non-goals

- Замена Jenkins на Tekton Controller или CRD-совместимый YAML.
- Tekton Triggers / EventListeners.
- Отдельный platform component pin для containerfile catalog (P0 — секция GP release body).
- Изменение GP composition (2-pin) или product config v2.
- Corp fleet migration prod GP.

## Impact

| Область | Изменения |
|---------|-----------|
| **coin-api** | schema v4 validate/preview; containerfile catalog storage; v3 read compat |
| **coin-ui** | pipeline canvas, task editor, containerfiles panel |
| **coin-executor** | step dispatch; containerfile by ref |
| **coin-lib** | `coinPipeline` — task DAG → Jenkins stages |
| **coin-gp-content** | seed templates v4 shape |
| **docs/adr** | ADR pipeline-tekton-mapping |
| **openspec/specs** | 2 new + 7 modified capabilities |
