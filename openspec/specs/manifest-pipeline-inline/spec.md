# manifest-pipeline-inline Specification

## Purpose

Resolved manifest shape for pipeline-inline v3: inline steps, per-step containerfile refs, no top-level catalogs.

## Requirements

### Requirement: Pipeline-inline resolved manifest

Resolved manifest SHALL contain `parameters`, `validateSchema` ref, and `pipeline.stages` with inline steps materialized from the embedded GP release pipeline body merged with agent and branching-model pins. Buildkit steps in manifest MUST include co-located `containerfile` with `contentRef` and `digest` on the same step object. Manifest MUST NOT contain top-level `build.targets`, `deliverables`, or `artifacts.containerfiles`.

#### Scenario: Materialize go-app v3 manifest from GP release

- **WHEN** coin-api resolves GP release `go-app@1.0.0` with embedded pipeline-inline body
- **THEN** each buildkit run/build step in manifest MUST include inline step config plus `containerfile.contentRef` on that step
- **AND** manifest MUST NOT include separate `artifacts.containerfiles` array
- **AND** MUST NOT require loading a `gp-content` component package

#### Scenario: Manifest hash covers embedded pipeline step containerfile

- **WHEN** publisher changes `containerfile.body` on a pipeline step in GP draft
- **THEN** manifest hash MUST change
- **AND** resolve output MUST update that step's containerfile ref

### Requirement: Manifest self-sufficiency for Nexus fallback

Pipeline-inline manifest SHALL be sufficient for coin-executor without live PostgreSQL. Executor MUST obtain Containerfile materialization path from the step being executed.

#### Scenario: Nexus fallback resolve

- **WHEN** Jenkins resolves manifest from Nexus fallback
- **THEN** manifest MUST contain per-step containerfile content refs needed for buildkit dispatch
- **AND** MUST NOT require top-level containerfile catalog lookup

### Requirement: Publish destinations compatibility

Inline publish steps SHALL materialize publish metadata required by coin-executor without reintroducing deliverables catalog.

#### Scenario: Publish image step materialization

- **WHEN** pipeline includes publish step for prior build step of type `image`
- **THEN** executor MUST resolve publish routing from inline build metadata on referenced step
