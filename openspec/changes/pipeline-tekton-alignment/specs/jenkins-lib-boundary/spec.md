## ADDED Requirements

### Requirement: Jenkins pipeline expands task DAG

coin-lib `coinPipeline` SHALL expand manifest `pipeline.tasks` into Jenkins stages using topological order of `runAfter`. Each Jenkins stage MUST invoke `coinRunStage` with task `id` matching manifest `pipeline.tasks[].id`.

#### Scenario: Linear runAfter chain

- **WHEN** manifest defines tasks `validate`, `test`, `build`, `publish` with linear `runAfter`
- **THEN** coinPipeline MUST emit Jenkins stages in DAG order
- **AND** each stage MUST call coin-executor with `--task` equal to task id

#### Scenario: Default order without runAfter

- **WHEN** manifest tasks omit `runAfter` but appear in author order
- **THEN** coinPipeline MUST treat author order as sequential dependencies for P0

#### Scenario: No Groovy build logic in stage expansion

- **WHEN** coinPipeline expands tasks to Jenkins stages
- **THEN** Groovy MUST NOT interpret build, containerfile, or publish logic
- **AND** MUST delegate all step execution to coin-executor
