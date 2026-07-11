## MODIFIED Requirements

### Requirement: Build stacks catalog

The coin-ui SHALL provide a Platform → Build stacks catalog for all `gp-content` components.

#### Scenario: List gp-content stacks

- **WHEN** enabling team opens `/platform/build-stacks`
- **THEN** the UI MUST list all `gp-content` components from the registry with versions grouped by status (`draft`, `published`)
- **AND** MUST NOT filter or align stacks by GP profile name

#### Scenario: Open stack editor from catalog

- **WHEN** enabling team selects a `draft` gp-content version from the catalog
- **THEN** the UI MUST navigate to `/platform/build-stacks/{name}/{version}/edit`
- **AND** MUST NOT link to `/studio`

#### Scenario: Open published stack detail

- **WHEN** enabling team selects a `published` gp-content version from the catalog
- **THEN** the UI MUST navigate to `/platform/build-stacks/{name}/{version}` as read-only detail

#### Scenario: Create new build stack draft

- **WHEN** enabling team with publisher role clicks create new build stack from the catalog
- **THEN** the UI MUST create a draft via Admin API and open the edit page under `/platform/build-stacks`

## MODIFIED Requirements

### Requirement: gp-content from GP release composition

When viewing a GP release, the UI SHALL surface the pinned gp-content from that release's composition — not from the profile entity.

#### Scenario: Platform link from release composition

- **WHEN** enabling team views release detail and the composition includes `gp-content/go-app@1.0.0`
- **THEN** the UI MUST offer a link to `/platform/build-stacks/go-app/1.0.0` or edit route for draft versions
- **AND** MUST NOT link to `/studio`
- **AND** MUST NOT link to a profile-level Build stack hub tab
