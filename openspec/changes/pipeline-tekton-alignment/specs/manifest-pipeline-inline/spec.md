## MODIFIED Requirements

### Requirement: Pipeline-inline resolved manifest

Resolved manifest SHALL contain `parameters`, `validateSchema` ref, `containerfiles[]` with `contentRef` and `digest` for managed entries, and `pipeline.tasks` materialized from the embedded GP release pipeline body merged with agent and branching-model pins. Containerfile steps in manifest MUST resolve catalog refs to materialized entries. Manifest MUST NOT contain top-level `build.targets`, `deliverables`, or legacy `artifacts.containerfiles`. Manifest MUST NOT use `pipeline.stages` for v4 GP releases.

#### Scenario: Materialize go-app v4 manifest from GP release

- **WHEN** coin-api resolves GP release `go-app@1.0.0` with embedded pipeline-inline v4 body
- **THEN** manifest MUST include `containerfiles[]` with content refs for managed entries
- **AND** each `kind: containerfile` step MUST reference resolved catalog entry
- **AND** MUST NOT require loading a `gp-content` component package

#### Scenario: Manifest hash covers catalog content

- **WHEN** publisher changes managed `containerfiles[].body` on GP draft
- **THEN** manifest hash MUST change
- **AND** resolve output MUST update catalog digest and dependent steps

### Requirement: Manifest self-sufficiency for Nexus fallback

Pipeline-inline manifest SHALL be sufficient for coin-executor without live PostgreSQL. Executor MUST obtain Containerfile materialization from manifest `containerfiles` catalog and step refs.

#### Scenario: Nexus fallback resolve

- **WHEN** Jenkins resolves manifest from Nexus fallback
- **THEN** manifest MUST contain `containerfiles` content refs needed for containerfile step dispatch
- **AND** MUST NOT require live GP draft lookup

### Requirement: Publish destinations compatibility

Coin publish steps SHALL materialize publish metadata required by coin-executor referencing `buildTaskId` on prior build task without reintroducing deliverables catalog.

#### Scenario: Publish image step materialization

- **WHEN** pipeline includes publish step for prior build task of type `image`
- **THEN** executor MUST resolve publish routing from build metadata on referenced task
