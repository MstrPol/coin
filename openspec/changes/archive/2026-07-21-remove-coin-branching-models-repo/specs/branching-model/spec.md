## ADDED Requirements

### Requirement: No git reference catalog for branching models

The system SHALL treat Platform registry (PostgreSQL draft bodies and published Nexus packages) as the only source of truth for branching-model content. A separate git tree `coin-branching-models/` MUST NOT be required for authoring, resolve, or production publish.

#### Scenario: Authoring without git catalog

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
- **THEN** they MUST be able to open the JSON Schema document without navigating to a removed git catalog
