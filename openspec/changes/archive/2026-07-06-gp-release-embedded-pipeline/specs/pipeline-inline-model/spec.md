## MODIFIED Requirements

### Requirement: Pipeline-inline canonical model

Pipeline-inline SHALL use `schemaVersion: 3` where build, publish and containerfile configuration lives only in `pipeline.stages[].steps[]`. The structured model SHALL be stored as the embedded body of a GP release draft, not as a separate `gp-content` component package. Author-facing model MUST NOT include top-level `build.targets`, `deliverables`, or `artifacts.containerfiles`.

#### Scenario: Valid v3 GP release pipeline shape

- **WHEN** publisher saves pipeline-inline model on GP draft `go-app@1.0.0-snapshot.1`
- **THEN** coin-api MUST persist the structured body on that GP release without separate targets, deliverables or containerfile catalogs
- **AND** MUST NOT create a `gp-content` component version

#### Scenario: Reject v2 catalog sections in GP pipeline draft

- **WHEN** GP release pipeline body contains `schemaVersion: 3` together with `build.targets`, `deliverables` or `artifacts.containerfiles`
- **THEN** validate MUST reject the draft
