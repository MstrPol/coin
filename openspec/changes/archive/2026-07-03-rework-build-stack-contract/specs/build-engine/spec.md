## MODIFIED Requirements

### Requirement: Build engine in manifest

Manifest SHALL include build targets where each target declares its own engine with value `buildkit` or `dockerfile`. A single top-level `build.engine` SHALL NOT be the dispatch source for Build Stack vNext.

#### Scenario: Resolve emits target engines

- **WHEN** product resolves GP with Build Stack vNext defining multiple targets
- **THEN** manifest MUST contain matching `build.targets[]` entries with per-target `engine`
- **AND** executor MUST dispatch each target using that target engine

### Requirement: Typed pipeline stages

Pipeline stages SHALL be typed executor stages with platform action steps. Stage steps SHALL reference parameters, targets or deliverables; they SHALL NOT contain product-provided shell scripts.

#### Scenario: Stage execution via executor

- **WHEN** Jenkins runs Test stage
- **THEN** coin-lib MUST invoke `coin-executor run --stage test`
- **AND** executor MUST evaluate manifest `pipeline.stages[]` steps for `test`
- **AND** executor MUST NOT execute arbitrary shell script from product config

### Requirement: BYO dockerfile engine

For a target with engine `dockerfile`, coin-executor SHALL build from Dockerfile path in product workspace per Build Stack target policy. coin-executor MUST NOT materialize managed Containerfile from gp-content package for that target.

#### Scenario: BYO target build from checkout

- **WHEN** executor runs target `app-image` with engine `dockerfile` and Dockerfile path `Dockerfile`
- **THEN** executor MUST use workspace `Dockerfile` as build definition
- **AND** MUST NOT write managed `.coin/Containerfile` for that target

### Requirement: Artifact deliverable target validation

Build Stack vNext SHALL validate artifact deliverables against the referenced target capabilities instead of applying a global engine restriction.

#### Scenario: Accept artifact with artifact-capable target

- **WHEN** draft Build Stack has artifact deliverable referencing target that can produce artifact output
- **THEN** validate-package MUST accept the deliverable

#### Scenario: Reject artifact on incompatible target

- **WHEN** draft Build Stack has artifact deliverable referencing target that cannot produce artifact output
- **THEN** validate-package MUST fail validation
