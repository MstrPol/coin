## ADDED Requirements

### Requirement: Materialize v4 resolved shape

coin-api manifest builder SHALL materialize a resolved manifest with `schemaVersion: 4` containing top-level `pipeline.tasks`, `containerfiles[]`, and `destinations[]` catalog matching the file-fixture contract from `pipeline-tekton-alignment`. Materialized output MUST NOT include `pipeline.stages`, top-level `build.targets`, `deliverables`, or `capabilities.deliverables`.

#### Scenario: Builder matches fixture shape

- **WHEN** coin-api materializes a seeded go-app v4 GP release
- **THEN** resolved JSON MUST expose `pipeline.tasks` and `containerfiles` with the same field semantics as `.coin/manifest.local.yaml`
- **AND** MUST be loadable by coin-executor ValidateV4 without adapter

### Requirement: Validate schemaVersion 4 on storage

coin-api SHALL reject GP draft/release bodies that claim schemaVersion 4 but fail `pipeline-inline.v4.schema.json` / manifest v4 rules (tasks, catalog ids, destination refs).

#### Scenario: Invalid v4 draft rejected

- **WHEN** draft body has `schemaVersion: 4` and an unknown `destinationRef` or missing `pipeline.tasks`
- **THEN** API MUST reject save with a validation error

### Requirement: Preview returns v4 resolved preview

Preview API SHALL return the same resolved shape fields that Jenkins consume after remote resolve (tasks, containerfiles, destinations), suitable for UI preview without calling Nexus for the structural check.

#### Scenario: Preview includes tasks and catalog

- **WHEN** enabling user requests preview for a v4 draft
- **THEN** response MUST include `pipeline.tasks` and `containerfiles[]`
