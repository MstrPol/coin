# branching-model Specification

## Purpose
TBD - created by archiving change gp-branching-model. Update Purpose after archive.
## Requirements
### Requirement: Branching model component type

The system SHALL register `branching-model` as a GP composition component type with semver versions using `draft` and `published` statuses only; immutable Nexus packages exist only for `published` versions.

#### Scenario: Publish draft model to stable

- **WHEN** enabling team publishes `branching-model/trunk-based@1.0.0` from Platform UI
- **THEN** the version MUST transition from `draft` to `published`
- **AND** MUST have an immutable Nexus package containing `model.yaml` and a full content_ref v2 with package reference

#### Scenario: No component canary status

- **WHEN** enabling team attempts to set branching model version status to `canary`
- **THEN** coin-api MUST reject the request

### Requirement: Five-slot GP composition

GP composition SHALL include a `branching-model` slot pinning model name and version.

#### Scenario: Resolve branching from draft pin on canary

- **WHEN** product CI resolves GP on canary channel with pinned branching-model version in `draft` status
- **THEN** manifest MUST include `branching` section materialized from PostgreSQL draft bodies

#### Scenario: Resolve branching from published pin on stable

- **WHEN** product CI resolves GP on stable channel with pinned branching-model in `published` status
- **THEN** manifest MUST include `branching` section materializable from Nexus package (with PG as admin fallback)

### Requirement: Model schema

Each branching model SHALL conform to `branching-model.schema.json` defining trunk, branch naming, versioning, and publish policy.

#### Scenario: Invalid model rejected

- **WHEN** model.yaml fails schema validation
- **THEN** publish MUST be rejected before registry write

