# branching-model Specification

## Purpose
TBD - created by archiving change gp-branching-model. Update Purpose after archive.
## Requirements
### Requirement: Branching model component type

The system SHALL register `branching-model` as a GP composition component type with semver versions; immutable Nexus packages exist only for `published` versions.

#### Scenario: Publish model version to canary

- **WHEN** enabling team publishes `branching-model/trunk-based@1.0.0` to canary via Component Studio
- **THEN** the version MUST be stored in component registry with `model.yaml` in PostgreSQL and MUST NOT create an immutable Nexus package

#### Scenario: Promote model version to stable

- **WHEN** enabling team promotes `branching-model/trunk-based@1.0.0` to `published`
- **THEN** the version MUST have an immutable Nexus package containing `model.yaml` and a full content_ref v2 with package reference

### Requirement: Five-slot GP composition

GP composition SHALL include a `branching-model` slot pinning model name and version.

#### Scenario: Resolve five slots from canary

- **WHEN** product CI resolves GP release with 5-slot composition on canary channel and pinned branching-model is `canary`
- **THEN** manifest MUST include `branching` section materialized from PostgreSQL without requiring Nexus package

#### Scenario: Resolve five slots from published

- **WHEN** product CI resolves GP release with 5-slot composition on stable channel and pinned branching-model is `published`
- **THEN** manifest MUST include `branching` section materializable from Nexus package (with PG as admin fallback)

### Requirement: Model schema

Each branching model SHALL conform to `branching-model.schema.json` defining trunk, branch naming, versioning, and publish policy.

#### Scenario: Invalid model rejected

- **WHEN** model.yaml fails schema validation
- **THEN** publish MUST be rejected before registry write

