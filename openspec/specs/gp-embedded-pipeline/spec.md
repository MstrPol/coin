# gp-embedded-pipeline Specification

## Purpose

Embedded pipeline-inline model on GP release draft: storage, validation, preview, and manifest materialization on promote.

## Requirements

### Requirement: Pipeline-inline body on GP release draft

GP release draft SHALL store pipeline-inline `schemaVersion: 3` structured model as the primary release payload. Storage MUST include `parameters`, `validateSchema`, and `pipeline.stages` with inline steps. coin-api MUST NOT require a separate `gp-content` component pin to resolve pipeline for that release.

#### Scenario: Save pipeline on GP draft

- **WHEN** publisher saves pipeline-inline model on GP draft `go-app@1.0.0-snapshot.1`
- **THEN** coin-api MUST persist the structured body scoped to that GP release version
- **AND** MUST NOT create or update a `gp-content` component version

#### Scenario: Pipeline changes require GP version bump

- **WHEN** publisher needs to change pipeline for profile `go-app`
- **THEN** publisher MUST create a new GP release version (draft)
- **AND** coin-api MUST NOT allow mutating pipeline body on a published GP release

### Requirement: Pipeline validation on GP release

coin-api SHALL validate embedded pipeline-inline model with the same rules as former gp-content v3 validation (short hash ids, inline containerfile body on buildkit steps, publish graph, no catalog sections).

#### Scenario: Reject buildkit step without containerfile body

- **WHEN** GP draft pipeline contains buildkit `run` step without `containerfile.body`
- **THEN** validate MUST return field-level error scoped to GP release pipeline path

#### Scenario: Reject v2 catalog sections

- **WHEN** GP draft pipeline body contains `build.targets`, `deliverables`, or `artifacts.containerfiles`
- **THEN** validate MUST reject the draft

### Requirement: GP release pipeline preview API

coin-api SHALL expose preview endpoint scoped to GP release that accepts or reads pipeline-inline structured model and returns resolved manifest subset plus validation issues.

#### Scenario: Preview embedded pipeline draft

- **WHEN** publisher posts valid pipeline-inline model to GP release preview endpoint
- **THEN** coin-api MUST return resolved inline pipeline manifest fragment
- **AND** MUST include per-step containerfile content refs in preview output

#### Scenario: Preview rejects invalid publish graph

- **WHEN** preview request contains `publish` step with missing `buildStepId`
- **THEN** preview MUST return validation error with field path to the publish step

### Requirement: Published pipeline in manifest blob

On GP promote, coin-api SHALL materialize embedded pipeline into the canonical Nexus manifest blob for that GP release. Published manifest MUST contain `parameters`, `validateSchema`, and `pipeline` sections sourced from the GP release pipeline body.

#### Scenario: Promote embeds pipeline in manifest

- **WHEN** publisher promotes GP draft with valid embedded pipeline
- **THEN** coin-api MUST write manifest blob containing materialized pipeline sections
- **AND** `manifestHash` MUST reflect embedded pipeline content

#### Scenario: Resolve reads pipeline from manifest for published release

- **WHEN** product CI resolves published GP release `go-app@1.0.0`
- **THEN** coin-api MUST serve manifest with pipeline from published blob
- **AND** MUST NOT require lookup of `gp-content` component package
