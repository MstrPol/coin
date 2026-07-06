## MODIFIED Requirements

### Requirement: Pipeline-inline body on GP release draft

GP release draft SHALL store pipeline-inline `schemaVersion: 4` structured model as the primary release payload. Storage MUST include `parameters`, `validateSchema`, `containerfiles`, and `pipeline.tasks` with typed steps. coin-api MUST NOT require a separate `gp-content` component pin to resolve pipeline for that release. coin-api SHALL accept read-only v3 bodies and migrate to v4 on draft save.

#### Scenario: Save pipeline on GP draft

- **WHEN** publisher saves pipeline-inline v4 model on GP draft `go-app@1.0.0-snapshot.1`
- **THEN** coin-api MUST persist the structured body scoped to that GP release version
- **AND** MUST NOT create or update a `gp-content` component version

#### Scenario: Pipeline changes require GP version bump

- **WHEN** publisher needs to change pipeline for profile `go-app`
- **THEN** publisher MUST create a new GP release version (draft)
- **AND** coin-api MUST NOT allow mutating pipeline body on a published GP release

### Requirement: Pipeline validation on GP release

coin-api SHALL validate embedded pipeline-inline v4 model: semantic task ids, acyclic `runAfter`, catalog refs, coin/containerfile/sh step rules, and no legacy catalog sections. coin-api SHALL validate v3 bodies using legacy rules until migrated.

#### Scenario: Reject containerfile step without catalog ref

- **WHEN** GP draft pipeline contains `kind: containerfile` step with unknown `ref`
- **THEN** validate MUST return field-level error scoped to GP release pipeline path

#### Scenario: Reject v2 catalog sections

- **WHEN** GP draft pipeline body contains `build.targets`, `deliverables`, or legacy `artifacts.containerfiles`
- **THEN** validate MUST reject the draft

### Requirement: GP release pipeline preview API

coin-api SHALL expose preview endpoint scoped to GP release that accepts or reads pipeline-inline v4 structured model and returns resolved manifest subset plus validation issues including materialized `containerfiles` catalog.

#### Scenario: Preview embedded pipeline draft

- **WHEN** publisher posts valid pipeline-inline v4 model to GP release preview endpoint
- **THEN** coin-api MUST return resolved manifest fragment with `containerfiles` and `pipeline.tasks`
- **AND** MUST include per-catalog-entry `contentRef` and digest in preview output

#### Scenario: Preview rejects invalid publish graph

- **WHEN** preview request contains `coin` publish step with missing `buildTaskId`
- **THEN** preview MUST return validation error with field path to the publish step

### Requirement: Published pipeline in manifest blob

On GP promote, coin-api SHALL materialize embedded pipeline v4 into the canonical Nexus manifest blob for that GP release. Published manifest MUST contain `parameters`, `validateSchema`, `containerfiles`, and `pipeline` sections sourced from the GP release pipeline body.

#### Scenario: Promote embeds pipeline in manifest

- **WHEN** publisher promotes GP draft with valid embedded pipeline v4
- **THEN** coin-api MUST write manifest blob containing materialized `containerfiles` and `pipeline.tasks`
- **AND** `manifestHash` MUST reflect catalog and task content

#### Scenario: Resolve reads pipeline from manifest for published release

- **WHEN** product CI resolves published GP release `go-app@1.0.0`
- **THEN** coin-api MUST serve manifest with pipeline and containerfiles from published blob
- **AND** MUST NOT require lookup of `gp-content` component package
