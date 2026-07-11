## ADDED Requirements

### Requirement: GP content schema v2 editor

The build stack editor SHALL edit `content.yaml` schema v2 bijectively with ordered section cards: engine, build policy, capabilities, pipeline stages, and artifacts.

#### Scenario: Engine card switches build block

- **WHEN** publisher selects engine `dockerfile` in the editor
- **THEN** the UI MUST show BYO dockerfile fields (`path`, `imageTarget`, `testTarget`)
- **AND** MUST hide buildkit targets and managed containerfile artifact key

#### Scenario: Engine card buildkit

- **WHEN** publisher selects engine `buildkit`
- **THEN** the UI MUST show buildkit targets map and managed containerfile artifact editor
- **AND** MUST allow `artifact` in capabilities deliverables

#### Scenario: Save produces v2 yaml

- **WHEN** publisher saves draft from the editor
- **THEN** persisted `content.yaml` MUST have `schemaVersion: 2`
- **AND** MUST NOT contain `controls` or `pipeline.stages[].when`

### Requirement: Build stack preview panel

The build stack editor SHALL call gp-content preview API and display resolved manifest fragment and warnings.

#### Scenario: Debounced preview on edit

- **WHEN** publisher changes engine or targets in draft editor
- **THEN** the UI MUST call preview API and show resolved `build` snippet

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
