## ADDED Requirements

### Requirement: Tekton entity mapping documentation

Project documentation SHALL document Tekton-to-Coin entity mapping for authoring and Jenkins runtime. Mapping MUST state that Coin does NOT run Tekton Controller.

#### Scenario: ADR lists entity mapping

- **WHEN** reader opens pipeline Tekton mapping ADR
- **THEN** document MUST map Pipeline→GP/Jenkins pipeline, Task→Jenkins stage, Step→executor action, Workspace→agent checkout, PipelineRun→Jenkins build
- **AND** MUST state TaskRun and Triggers are out of scope

### Requirement: Pipeline-inline schemaVersion 4 canonical model

Pipeline-inline SHALL use `schemaVersion: 4` with `pipeline.tasks[]`, optional `runAfter` on each task, and top-level `containerfiles[]` catalog. Author model MUST NOT use `pipeline.stages[]` for new drafts. Author model MUST NOT duplicate managed Containerfile body inline on steps.

#### Scenario: Valid v4 GP release pipeline shape

- **WHEN** publisher saves pipeline-inline v4 model on GP draft `go-app@1.0.0-snapshot.1`
- **THEN** coin-api MUST persist `parameters`, `validateSchema`, `containerfiles`, and `pipeline.tasks`
- **AND** MUST reject body with `pipeline.stages` for schemaVersion 4

#### Scenario: Reject v2 catalog sections in v4 draft

- **WHEN** GP release pipeline body contains `schemaVersion: 4` together with `build.targets`, `deliverables`, or legacy `artifacts.containerfiles`
- **THEN** validate MUST reject the draft

### Requirement: Task graph with runAfter

Each pipeline task SHALL declare unique semantic `id` and MAY declare `runAfter` referencing other task ids. Validation MUST ensure `runAfter` references exist and the graph is acyclic.

#### Scenario: Linear task chain

- **WHEN** draft defines tasks `validate`, `test`, `build`, `publish` with `runAfter` chain
- **THEN** validate MUST accept the draft
- **AND** manifest materialization MUST preserve task order compatible with DAG

#### Scenario: Reject cyclic runAfter

- **WHEN** task A has `runAfter: [B]` and task B has `runAfter: [A]`
- **THEN** validate MUST reject the draft with cycle error

#### Scenario: Reject missing runAfter target

- **WHEN** task references `runAfter: [missing]`
- **THEN** validate MUST reject the draft

### Requirement: Semantic task ids

`pipeline.tasks[].id` SHALL match `^[a-z][a-z0-9-]{1,31}$`. Human-readable labels belong in `task.name`.

#### Scenario: Accept semantic task id

- **WHEN** task `id` is `validate` or `build-image`
- **THEN** validate MUST accept the id

#### Scenario: Reject short hash task id on v4 draft

- **WHEN** v4 draft uses task `id` matching only `^[a-z0-9]{5,6}$` without semantic meaning and migration flag is absent
- **THEN** validate MUST reject or require migration to semantic id

### Requirement: Typed pipeline steps v4

Each task step SHALL declare `kind` as `coin`, `containerfile`, or `sh` with exactly one matching config block.

#### Scenario: Coin validate step

- **WHEN** publisher adds step `kind: coin` with `action: validate`
- **THEN** validate MUST accept step without containerfile ref
- **AND** executor MUST dispatch validate primitive

#### Scenario: Containerfile step references catalog

- **WHEN** publisher adds step `kind: containerfile` with `ref: app`
- **THEN** step MUST reference existing `containerfiles[].id`
- **AND** validation MUST reject missing catalog ref

#### Scenario: Sh step allowlist

- **WHEN** publisher adds step `kind: sh` with `script` not on pilot allowlist
- **THEN** validate MUST reject the draft

### Requirement: v3 read compat and migration on save

coin-api SHALL read published v3 pipeline bodies for resolve. coin-api SHALL offer automatic v3→v4 migration when publisher saves GP draft pipeline body.

#### Scenario: Resolve published v3 manifest

- **WHEN** product resolves published GP release with embedded v3 pipeline
- **THEN** coin-api MUST return executable manifest without requiring manual migration

#### Scenario: Save draft migrates v3 to v4

- **WHEN** publisher saves v3 pipeline body on GP draft
- **THEN** coin-api MUST migrate `stages` to `tasks`, extract inline containerfiles to catalog, and persist `schemaVersion: 4`
