## MODIFIED Requirements

### Requirement: Build stacks catalog

The coin-ui SHALL provide a Platform → Build stacks catalog listing **gp-content profiles** (not a flat version list).

#### Scenario: List gp-content profiles

- **WHEN** enabling team opens `/platform/build-stacks`
- **THEN** the UI MUST list all `gp-content` component profiles with per-profile version summary
- **AND** MUST NOT filter or align stacks by GP profile name

#### Scenario: Open stack hub from catalog

- **WHEN** enabling team selects a gp-content profile from the catalog
- **THEN** the UI MUST navigate to `/platform/build-stacks/{name}`

#### Scenario: Open draft editor from hub

- **WHEN** enabling team selects a `draft` gp-content version from the hub Releases tab
- **THEN** the UI MUST navigate to `/platform/build-stacks/{name}/{version}/edit`
- **AND** MUST NOT link to `/studio`

#### Scenario: Open published stack detail

- **WHEN** enabling team selects a `published` gp-content version from the hub
- **THEN** the UI MUST navigate to `/platform/build-stacks/{name}/releases/{version}` as read-only detail

#### Scenario: Create new build stack profile

- **WHEN** enabling team with publisher role clicks «New profile» on the catalog
- **THEN** the UI MUST create a component profile and open hub at `/platform/build-stacks/{name}?welcome=1`

#### Scenario: Create new build stack draft from hub

- **WHEN** publisher clicks «New draft» on build stack hub
- **THEN** the UI MUST create a draft version via Admin API and open the edit page

### Requirement: gp-content from GP release composition

When viewing a GP release, the UI SHALL surface the pinned gp-content from that release's composition — not from the profile entity.

#### Scenario: Platform link from release composition

- **WHEN** enabling team views release detail and the composition includes `gp-content/go-app@1.0.0`
- **THEN** the UI MUST offer a link to `/platform/build-stacks/go-app/releases/1.0.0` or edit route for draft versions
- **AND** MUST NOT link to `/studio`
- **AND** MUST NOT link to a GP profile-level Build stack tab
