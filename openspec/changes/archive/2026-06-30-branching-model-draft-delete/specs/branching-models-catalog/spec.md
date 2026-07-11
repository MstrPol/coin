## ADDED Requirements

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

## MODIFIED Requirements

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
