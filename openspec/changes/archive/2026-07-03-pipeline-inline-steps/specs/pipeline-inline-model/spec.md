## ADDED Requirements

### Requirement: Pipeline-inline canonical model

Build Stack pipeline-inline SHALL use `schemaVersion: 3` where build, publish and containerfile configuration lives only in `pipeline.stages[].steps[]`. Author-facing model MUST NOT include top-level `build.targets`, `deliverables`, or `artifacts.containerfiles`.

#### Scenario: Valid v3 package shape

- **WHEN** publisher saves gp-content with `schemaVersion: 3`, parameters, validateSchema and pipeline stages with inline steps
- **THEN** coin-api MUST persist the structured package without separate targets, deliverables or containerfile catalogs

#### Scenario: Reject v2 catalog sections in v3 draft

- **WHEN** draft contains `schemaVersion: 3` together with `build.targets`, `deliverables` or `artifacts.containerfiles`
- **THEN** validate-package MUST reject the draft

### Requirement: Typed inline pipeline steps

Each pipeline step SHALL declare `action` as `run`, `build`, or `publish` with exactly one matching inline config block.

#### Scenario: Run step inline config

- **WHEN** publisher adds `action: run`
- **THEN** step MUST include `run` with `engine` and required engine fields
- **AND** validation MUST reject incomplete run config

#### Scenario: Build step inline config

- **WHEN** publisher adds `action: build`
- **THEN** step MUST include unique `build.id`, `type`, engine fields and type-specific metadata inline
- **AND** validation MUST reject duplicate `build.id`

#### Scenario: Publish step references build output

- **WHEN** publisher adds `action: publish`
- **THEN** step MUST include `publish.buildStepId` referencing existing `build.id`
- **AND** validation MUST reject missing `buildStepId`

### Requirement: Short hash stage and build ids

`pipeline.stages[].id` and `build.id` SHALL match `^[a-z0-9]{5,6}$` (5–6 lowercase alphanumeric characters).

#### Scenario: Reject semantic slug ids

- **WHEN** stage `id` is `validate` or `build.id` is `app-image`
- **THEN** validate-package MUST reject ids that do not match the short hash pattern


Buildkit `run` and `build` steps SHALL carry managed Containerfile content inline on the same step object. Author model MUST use `containerfile.body`. Author model MUST NOT reference a separate containerfile catalog by id.

#### Scenario: Buildkit step requires containerfile body

- **WHEN** inline step uses `engine: buildkit`
- **THEN** step MUST include non-empty `containerfile.body`
- **AND** validation MUST reject buildkit step without containerfile body

#### Scenario: Reject containerfile catalog reference

- **WHEN** v3 draft uses `containerfile: app` id reference without inline body
- **THEN** validate-package MUST reject the draft

### Requirement: BYO dockerfile inline in step

Dockerfile engine steps SHALL specify workspace dockerfile path inline on the step via `dockerfile.path` without managed containerfile body.

#### Scenario: Dockerfile engine step

- **WHEN** inline step uses `engine: dockerfile`
- **THEN** step MUST include `dockerfile.path`
- **AND** step MUST NOT require `containerfile.body`

### Requirement: Parameters unchanged from vNext

Pipeline-inline SHALL retain typed non-secret parameters with the same validation rules as Build Stack vNext.

#### Scenario: Enum parameter requires allowed values

- **WHEN** parameter has `type: enum`
- **THEN** validation MUST require non-empty `allowedValues`
