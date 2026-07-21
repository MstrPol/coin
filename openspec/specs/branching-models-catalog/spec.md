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

The branching models catalog SHALL display **branching-model profiles** with version summary, lifecycle status counts, and linked GP profiles.

#### Scenario: View profile list

- **WHEN** enabling team opens the branching models catalog
- **THEN** the UI MUST show for each model profile at least: name, draft and published version counts, and created/updated metadata when available

#### Scenario: View GP profile usage

- **WHEN** a branching model profile is referenced by a GP release composition
- **THEN** the catalog MUST list GP profile names that pin that model name

#### Scenario: Open model hub from catalog

- **WHEN** enabling team selects a branching model profile from the catalog
- **THEN** the UI MUST navigate to `/platform/branching-models/{name}`

#### Scenario: Open draft version editor

- **WHEN** enabling team selects a `draft` branching model version from the hub
- **THEN** the UI MUST navigate to `/platform/branching-models/{name}/{version}/edit`
- **AND** MUST NOT link to `/studio`

#### Scenario: Open published version detail

- **WHEN** enabling team selects a `published` branching model version from the hub
- **THEN** the UI MUST navigate to `/platform/branching-models/{name}/releases/{version}` as read-only detail

### Requirement: Catalog actions for lifecycle

The catalog and hub SHALL expose lifecycle actions appropriate to the version status without requiring navigation through the generic components page or Component Studio.

#### Scenario: Publish draft from hub

- **WHEN** a branching model version is in `draft` status and validation passes
- **THEN** the hub release detail MUST offer Publish transitioning the version to `published`

#### Scenario: Create new branching model profile

- **WHEN** enabling team clicks «New profile» on the branching models catalog
- **THEN** the UI MUST create a component profile and open hub at `/platform/branching-models/{name}?welcome=1`

#### Scenario: Create new model version from hub

- **WHEN** publisher clicks «New draft» on branching model hub
- **THEN** the UI MUST create a draft and open `/platform/branching-models/{name}/{version}/edit`
- **AND** MUST NOT route to Component Studio

#### Scenario: Delete orphan draft from hub

- **WHEN** publisher deletes a branching-model draft that is not needed
- **THEN** the hub MUST remove the version from the Releases list after successful Admin API delete
- **AND** MUST NOT require Component Studio or legacy `/components` routes

### Requirement: Branching model rule builder

The branching model editor SHALL provide an ordered list of branch rule cards editing schema v2 fields: `name`, `pattern`, `versioning.template`, and `publish`.

The editor SHALL NOT call a branching preview API or evaluate branch/version/publish scenarios in `coin-api`.

#### Scenario: Editor maps to model.yaml

- **WHEN** publisher saves a draft branching model from the editor
- **THEN** the persisted `model.yaml` MUST be valid schema v2 and reflect card order as `branches` list order

#### Scenario: Reorder branch rules

- **WHEN** publisher reorders branch cards in the editor
- **THEN** the YAML `branches` order MUST update (first match wins)

#### Scenario: Editor does not run platform preview

- **WHEN** publisher edits branch rule fields
- **THEN** the UI MUST NOT call `POST /v1/admin/branching-models/preview`
- **AND** MUST NOT show executor-backed scenario results

### Requirement: Branching model authoring documentation

The branching models catalog and editor SHALL link operators to `docs/how-to/branching-models.md` as the sole human authoring guide. The UI MUST NOT require or link to a separate tree under `coin-branching-models/`.

#### Scenario: How-to link from editor

- **WHEN** publisher opens branching model draft editor
- **THEN** the UI MUST provide a link or reference to the authoring how-to documentation
- **AND** MUST NOT depend on per-model README paths under `coin-branching-models/models/{name}/`

### Requirement: Delete branching-model draft from hub

The branching models hub SHALL allow publishers to delete draft versions from the Releases tab and the draft editor lifecycle panel.

#### Scenario: Delete draft from releases list

- **WHEN** publisher views `/platform/branching-models/{name}/releases`
- **AND** a version row has `status = draft`
- **THEN** the UI MUST offer a «Delete» or «Delete draft» action for that row
- **AND** MUST call `DELETE /v1/admin/components/branching-model/{name}/versions/{version}`
- **AND** MUST NOT offer delete for `published` rows

#### Scenario: Delete draft from editor lifecycle panel

- **WHEN** publisher opens draft editor at `/platform/branching-models/{name}/{version}/edit`
- **THEN** the lifecycle panel MUST offer «Delete draft» alongside Validate and Publish
- **AND** MUST require confirmation before calling the Admin API
- **AND** after successful delete MUST navigate to `/platform/branching-models/{name}/releases`

#### Scenario: Non-publisher cannot delete

- **WHEN** user without publisher role views branching-model draft releases or editor
- **THEN** the UI MUST NOT show delete actions

