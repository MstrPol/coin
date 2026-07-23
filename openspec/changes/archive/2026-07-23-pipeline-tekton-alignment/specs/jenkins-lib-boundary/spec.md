## ADDED Requirements

### Requirement: Jenkins pipeline expands task DAG

coin-lib `coinPipeline` SHALL expand manifest `pipeline.tasks` into Jenkins stages using topological order of `runAfter`. Each Jenkins stage MUST invoke `coinRunStage` with task `id` matching manifest `pipeline.tasks[].id`.

#### Scenario: Linear runAfter chain

- **WHEN** manifest defines tasks `validate`, `test`, `build`, `publish` with linear `runAfter`
- **THEN** coinPipeline MUST emit Jenkins stages in DAG order
- **AND** each stage MUST call coin-executor with `--task` equal to task id

#### Scenario: Default order without runAfter

- **WHEN** manifest tasks omit `runAfter` but appear in author order
- **THEN** coinPipeline MUST treat author order as sequential dependencies for Phase A

#### Scenario: No Groovy build logic in stage expansion

- **WHEN** coinPipeline expands tasks to Jenkins stages
- **THEN** Groovy MUST NOT interpret build, containerfile, or publish logic
- **AND** MUST delegate all step execution to coin-executor

### Requirement: File-based manifest resolve in coin-lib

`coinResolveManifest` SHALL support `coin.resolve: file` per `config-resolve-file`. Remote path (API → Nexus) MUST remain the default when `resolve` is omitted or `remote`.

#### Scenario: Skip network on file resolve

- **WHEN** product config sets `coin.resolve: file` and fixture exists
- **THEN** coin-lib MUST NOT call coin-api or Nexus for that resolve
- **AND** MUST return the fixture map for materialize

#### Scenario: Remote resolve unchanged

- **WHEN** product config uses default or `coin.resolve: remote`
- **THEN** coin-lib MUST use coin-api primary and Nexus fallback as before

### Requirement: Archive test results from workspace

After a test task (or at end of pipeline when `.coin/test-results/` exists), coin-lib SHALL archive `.coin/test-results/**` as Jenkins build artifacts with `allowEmptyArchive` true. coin-lib MAY publish JUnit XML from that directory when present. Groovy MUST NOT run product tests; it only archives files produced by coin-executor.

#### Scenario: Archive test-results directory

- **WHEN** executor has written files under `.coin/test-results/`
- **THEN** coinPipeline or post-test glue MUST call Jenkins `archiveArtifacts` for `.coin/test-results/**`

#### Scenario: Empty test-results does not fail archive

- **WHEN** `.coin/test-results/` is missing or empty
- **THEN** archive step MUST NOT fail the build solely due to missing artifacts

### Requirement: Registry auth from destinations catalog

`coinConfigureRegistryAuth` SHALL configure Docker registry credentials for every host listed in schemaVersion 4 `destinations[]` where `pull` or `push` is true. Host MUST come from entry `host` when set, otherwise from `imageRegistryPrefix`. Legacy flat `destinations.imageRegistryPrefix` MUST remain supported for pre-v4 manifests.

#### Scenario: Auth covers pull and push destinations

- **WHEN** manifest destinations catalog has entry `nexus-docker` with `pull: true` and `push: true`
- **THEN** auth config MUST include that registry host before build/publish stages

#### Scenario: Legacy flat destinations still work

- **WHEN** manifest has flat `destinations.imageRegistryPrefix`
- **THEN** auth MUST configure the host derived from that prefix
