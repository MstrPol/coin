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

Each branching model SHALL conform to `branching-model.schema.json` **schemaVersion 2** with an ordered `branches` list. Each branch rule MUST define `name`, `pattern` (RE2), `versioning.template` (mini-DSL), and `publish` (boolean eligibility).

#### Scenario: Invalid model rejected

- **WHEN** model.yaml fails schema v2 validation or template placeholder rules
- **THEN** validate-package MUST reject before registry write

#### Scenario: Schema v1 rejected

- **WHEN** model.yaml has `schemaVersion: 1` or v1-only fields (`trunk`, `branchTypes`, `publish.when`)
- **THEN** coin-api MUST reject the draft

#### Scenario: Main branch rule present

- **WHEN** a valid v2 model is validated
- **THEN** it MUST include a branch rule matching `main` or `master` (e.g. `pattern: ^main$|^master$`)

### Requirement: No reference tree for branching models

The system SHALL treat Platform registry (PostgreSQL draft bodies and published Nexus packages) as the only source of truth for branching-model content. A tree `coin-branching-models/` MUST NOT be required for authoring, resolve, or production publish.

#### Scenario: Authoring without reference tree

- **WHEN** enabling team creates or updates a branching-model version via Platform UI
- **THEN** the workflow MUST complete using Admin API artifacts only
- **AND** MUST NOT require files from `coin-branching-models/`

#### Scenario: Local pilot seed uses testdata fixtures

- **WHEN** local docker seed publishes `branching-model/trunk-based` for pilot stacks
- **THEN** the seed MUST read `model.yaml` from `docker/testdata/branching-models/` (or equivalent fixture under `docker/`)
- **AND** MUST NOT invoke scripts from a `coin-branching-models/` directory

### Requirement: Schema documentation location

Documentation of branching-model schema v2 SHALL live at `docs/schemas/branching-model.schema.json` (or an equivalent path under `docs/`). Runtime validation MUST continue to enforce schemaVersion 2 rules in coin-api regardless of the documentation file location.

#### Scenario: Operators find schema next to how-to

- **WHEN** an operator follows links from `docs/how-to/branching-models.md`
- **THEN** they MUST be able to open the JSON Schema document without navigating to a removed `coin-branching-models/` tree

