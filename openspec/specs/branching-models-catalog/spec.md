# branching-models-catalog Specification

## Purpose

Branching models catalog in coin-ui under Platform: lifecycle statuses, GP usage, Platform editor links.
## Requirements
### Requirement: Branching models navigation entry

The coin-ui SHALL provide branching models under the Platform sidebar group, not as a top-level peer of Fleet and Golden Paths.

#### Scenario: Open branching models from Platform

- **WHEN** enabling team clicks "Branching models" under Platform in the sidebar
- **THEN** the UI MUST navigate to `/platform/branching-models` showing only `branching-model` components

#### Scenario: Legacy URL redirect

- **WHEN** user opens `/branching-models`
- **THEN** the UI MUST redirect to `/platform/branching-models`

### Requirement: Branching models catalog view

The branching models catalog SHALL display model name, version lines, lifecycle status per version, and linked GP profiles.

#### Scenario: View model statuses

- **WHEN** enabling team opens the branching models catalog
- **THEN** the UI MUST show for each model at least: name, versions grouped by status (`draft`, `published`), and created/updated metadata when available

#### Scenario: View GP profile usage

- **WHEN** a branching model is referenced by a GP release composition
- **THEN** the catalog MUST list GP profile names that pin that model name

#### Scenario: Open draft version editor

- **WHEN** enabling team selects a `draft` branching model version from the catalog
- **THEN** the UI MUST navigate to `/platform/branching-models/{name}/{version}/edit`
- **AND** MUST NOT link to `/studio`

#### Scenario: Open published version detail

- **WHEN** enabling team selects a `published` branching model version from the catalog
- **THEN** the UI MUST navigate to `/platform/branching-models/{name}/{version}` as read-only detail

### Requirement: Catalog actions for lifecycle

The catalog SHALL expose lifecycle actions appropriate to the version status without requiring navigation through the generic components page or Component Studio.

#### Scenario: Publish draft from catalog

- **WHEN** a branching model version is in `draft` status and validation passes
- **THEN** the catalog or detail page MUST offer Publish transitioning the version to `published`

#### Scenario: Create new model version

- **WHEN** enabling team starts a new branching model version from the catalog
- **THEN** the UI MUST create a draft and open `/platform/branching-models/{name}/{version}/edit`
- **AND** MUST NOT route to Component Studio

