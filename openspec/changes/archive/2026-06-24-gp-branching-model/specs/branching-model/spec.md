# branching-model

## ADDED Requirements

### Requirement: Branching model component type

The system SHALL register `branching-model` as a GP composition component type with semver versions published to Nexus.

#### Scenario: Publish model version

- **WHEN** enabling team publishes `branching-model/trunk-based@1.0.0` via Component Studio
- **THEN** the version MUST be stored in component registry with immutable package in Nexus containing `model.yaml`

### Requirement: Five-slot GP composition

GP composition SHALL include a `branching-model` slot pinning model name and version.

#### Scenario: Resolve five slots

- **WHEN** product CI resolves GP release with 5-slot composition
- **THEN** manifest MUST include `branching` section materialized from pinned branching-model version

### Requirement: Model schema

Each branching model SHALL conform to `branching-model.schema.json` defining trunk, branch naming, versioning, and publish policy.

#### Scenario: Invalid model rejected

- **WHEN** model.yaml fails schema validation
- **THEN** publish MUST be rejected before registry write
