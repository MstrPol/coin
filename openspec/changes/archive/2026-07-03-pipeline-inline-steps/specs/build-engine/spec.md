## MODIFIED Requirements

### Requirement: Build engine in manifest

Manifest SHALL derive build engine dispatch from inline pipeline step `run.engine` and `build.engine` fields. Manifest MUST NOT require top-level `build.engine` or `build.targets` catalog for pipeline-inline stacks.

#### Scenario: Resolve emits per-step engines

- **WHEN** product resolves GP with pipeline-inline gp-content containing buildkit and dockerfile steps
- **THEN** manifest pipeline steps MUST retain per-step `engine` values for executor dispatch
- **AND** manifest MUST NOT require top-level `build.engine` as sole dispatch source

### Requirement: Typed pipeline stages

Pipeline stages SHALL remain typed executor stage names. Each stage executes inline steps (`run`, `build`, `publish`) without GP shell scripts or Jenkins Shared Library business logic.

#### Scenario: Stage execution via inline steps

- **WHEN** Jenkins runs Test stage from manifest with inline `run` step
- **THEN** coin-executor MUST dispatch the run step using inline engine config
- **AND** coin-lib MUST invoke executor without interpreting build logic in Groovy

## ADDED Requirements

### Requirement: Inline step dispatch in executor

coin-executor SHALL execute pipeline-inline steps directly from manifest without requiring separate manifest `build.targets` or `deliverables` sections.

#### Scenario: Execute run step

- **WHEN** executor runs stage step with `action: run`, `run.engine: buildkit` and step-local `containerfile.contentRef`
- **THEN** executor MUST materialize containerfile from that step ref and run the configured target
- **AND** MUST NOT require top-level `artifacts.containerfiles` catalog

#### Scenario: Execute build then publish

- **WHEN** stage contains `build` step with `build.id: app` followed later by `publish` with `buildStepId: app`
- **THEN** executor MUST build output during build step
- **AND** publish step MUST publish the output associated with `build.id: app`

## REMOVED Requirements

### Requirement: Artifact deliverable buildkit only

**Reason**: Deliverable typing moves to inline `build.type`; validation enforced in v3 schema per step instead of v2 `capabilities.deliverables` + top-level engine rule.
**Migration**: Express artifact outputs as `build` steps with `type: artifact`; BYO dockerfile steps validated per inline engine fields.
