# branching-models-catalog Specification

## Purpose
TBD - created by archiving change branching-model-lifecycle. Update Purpose after archive.
## Requirements
### Requirement: Branching models navigation entry

The coin-ui SHALL provide a dedicated navigation item and route for branching models separate from the generic components list.

#### Scenario: Open branching models catalog

- **WHEN** enabling team clicks "Branching Models" in the main navigation
- **THEN** the UI MUST navigate to `/branching-models` showing only `branching-model` components

### Requirement: Branching models catalog view

The branching models catalog SHALL display model name, version lines, lifecycle status per version, and linked GP profiles.

#### Scenario: View model statuses

- **WHEN** enabling team opens the branching models catalog
- **THEN** the UI MUST show for each model at least: name, versions grouped by status (`draft`, `canary`, `published`), and created/updated metadata when available

#### Scenario: View GP profile usage

- **WHEN** a branching model is referenced by a GP profile slot
- **THEN** the catalog MUST list GP profile names that pin that model name

#### Scenario: Open version in Studio

- **WHEN** enabling team selects a model version from the catalog
- **THEN** the UI MUST provide a link to Component Studio at `/studio/branching-model/{name}/{version}`

### Requirement: Catalog actions for lifecycle

The catalog SHALL expose lifecycle actions appropriate to the version status without requiring navigation through the generic components page.

#### Scenario: Promote canary from catalog

- **WHEN** a version is in `canary` status and health gate criteria are met
- **THEN** the catalog MUST offer Promote to stable linking to the existing promote wizard flow

#### Scenario: Create new model version

- **WHEN** enabling team starts a new branching model version from the catalog
- **THEN** the UI MUST route to Component Studio create flow for `branching-model`

