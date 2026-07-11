## MODIFIED Requirements

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
