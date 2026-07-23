# pipeline-tekton-model Specification

## Purpose

schemaVersion 4 pipeline model aligned with Tekton concepts (documentation only — Coin does not run Tekton Controller): `pipeline.tasks`, `runAfter`, typed steps, top-level `containerfiles` catalog.

## Requirements

### Requirement: Tekton entity mapping documentation

Project documentation SHALL document Tekton-to-Coin entity mapping for authoring and Jenkins runtime. Mapping MUST state that Coin does NOT run Tekton Controller.

#### Scenario: ADR lists entity mapping

- **WHEN** reader opens pipeline Tekton mapping ADR
- **THEN** document MUST map Pipeline→GP/Jenkins pipeline, Task→Jenkins stage, Step→executor action, Workspace→agent checkout, PipelineRun→Jenkins build
- **AND** MUST state TaskRun and Triggers are out of scope

### Requirement: Pipeline-inline schemaVersion 4 canonical model

Pipeline-inline and resolved manifests for v4 SHALL use `schemaVersion: 4` with `pipeline.tasks[]`, optional `runAfter` on each task, and top-level `containerfiles[]` catalog. New v4 documents MUST NOT use `pipeline.stages[]`. Managed Containerfile body MUST NOT be duplicated inline on steps.

#### Scenario: Valid v4 resolved fixture shape

- **WHEN** platform team places a v4 resolved fixture at `.coin/manifest.local.yaml`
- **THEN** the document MUST include `pipeline.tasks` and MAY include `containerfiles`
- **AND** MUST NOT use `pipeline.stages` for schemaVersion 4

#### Scenario: Reject v2 catalog sections in v4 document

- **WHEN** a v4 document contains `build.targets`, `deliverables`, or legacy `artifacts.containerfiles`
- **THEN** schema validation MUST reject the document

### Requirement: No capabilities.deliverables on schemaVersion 4

v4 resolved and authoring pipeline documents MUST NOT use `capabilities.deliverables` to declare what the GP builds. Build outputs are defined by `pipeline.tasks` coin build steps (`build.type`) and recorded in `.coin/outputs.json`. Legacy `capabilities.deliverables` is superseded for schemaVersion 4.

#### Scenario: v4 fixture omits capabilities.deliverables

- **WHEN** a schemaVersion 4 file-resolve fixture or resolved manifest is authored
- **THEN** the document MUST NOT require `capabilities.deliverables`
- **AND** validation MUST accept a document without the `capabilities` section when `pipeline.tasks` is present

#### Scenario: Reject relying on capabilities for v4 build selection

- **WHEN** a v4 pipeline has coin build steps
- **THEN** executor MUST select build behavior from step `build.type` / catalog refs
- **AND** MUST NOT require `capabilities.deliverables` to run those tasks

### Requirement: Task graph with runAfter

Each pipeline task SHALL declare unique semantic `id` and MAY declare `runAfter` referencing other task ids. Validation MUST ensure `runAfter` references exist and the graph is acyclic.

#### Scenario: Linear task chain

- **WHEN** fixture defines tasks `validate`, `test`, `build`, `publish` with `runAfter` chain
- **THEN** validation MUST accept the document
- **AND** coin-lib MUST expand stages in DAG-compatible order

#### Scenario: Reject cyclic runAfter

- **WHEN** task A has `runAfter: [B]` and task B has `runAfter: [A]`
- **THEN** validation MUST reject the document with cycle error

#### Scenario: Reject missing runAfter target

- **WHEN** task references `runAfter: [missing]`
- **THEN** validation MUST reject the document

### Requirement: Semantic task ids

`pipeline.tasks[].id` SHALL match `^[a-z][a-z0-9-]{1,31}$`. Human-readable labels belong in `task.name`.

#### Scenario: Accept semantic task id

- **WHEN** task `id` is `validate` or `build-image`
- **THEN** validation MUST accept the id

### Requirement: Typed pipeline steps v4

Each task step SHALL declare `kind` as `coin`, `containerfile`, or `sh` with matching config.

#### Scenario: Coin validate step

- **WHEN** step is `kind: coin` with `action: validate`
- **THEN** validation MUST accept the step without containerfile ref
- **AND** executor MUST dispatch validate primitive

#### Scenario: Containerfile step references catalog

- **WHEN** step is `kind: containerfile` with `ref: app`
- **THEN** step MUST reference existing `containerfiles[].id`
- **AND** validation MUST reject missing catalog ref

#### Scenario: Coin build references catalog id only

- **WHEN** step is `kind: coin` with `action: build` and `containerfileRef: app`
- **THEN** validation MUST require `app` in top-level `containerfiles`
- **AND** MUST NOT require literal `imageRef` on the step

#### Scenario: Coin publish references build task id

- **WHEN** step is `kind: coin` with `action: publish` and `buildTaskId: build-app`
- **THEN** validation MUST require a task `build-app` in `pipeline.tasks`
- **AND** MUST NOT require `containerfileRef` on the publish step

#### Scenario: Test step references catalog target

- **WHEN** step is `kind: containerfile` with `ref: app` and `run.target: test`, or `kind: coin` with `action: test` and `containerfileRef: app`
- **THEN** validation MUST require catalog id `app`
- **AND** MUST treat this as the default unit-test path (not agent shell)

#### Scenario: Sh step allowlist

- **WHEN** step is `kind: sh` with `script` not on pilot allowlist
- **THEN** validation or executor MUST reject before execution
