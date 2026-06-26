# platform-runtime-catalog Specification

## Purpose

Platform catalog for runtime components (agent, executor) in coin-ui.

## Requirements

### Requirement: Runtime catalog page

The coin-ui SHALL provide a Platform → Runtime catalog listing `agent` and `executor` components only.

The page MUST NOT display a platform lib pin banner, «Current platform lib pin» heading, link to edit lib in Platform settings, or `lib` component rows.

#### Scenario: No lib pin banner on runtime page

- **WHEN** enabling team opens `/platform/runtime`
- **THEN** the UI MUST NOT show text «Current platform lib pin» or any lib version pin block
- **AND** MUST NOT fetch platform settings solely to display a lib pin

#### Scenario: List runtime components

- **WHEN** enabling team opens `/platform/runtime`
- **THEN** the UI MUST show only components with type `agent` or `executor`, with name, version lines, and lifecycle status
- **AND** MUST NOT show type `lib` components

#### Scenario: Open runtime component detail

- **WHEN** enabling team selects a runtime component row
- **THEN** the UI MUST navigate to the existing component detail route for that type and name

### Requirement: Runtime publish guidance

The runtime catalog SHALL surface script-first publish path without in-app draft editors.

#### Scenario: Show publish hint

- **WHEN** enabling team views runtime catalog or component detail
- **THEN** the catalog MUST show a link or hint to the publish runbook for registering new agent or executor versions
- **AND** MUST NOT reference Component Studio

#### Scenario: No draft runtime versions in UI

- **WHEN** enabling team views `/platform/runtime`
- **THEN** the UI MUST list only `published` runtime versions for composition selection contexts
- **AND** MUST NOT offer create-draft or edit actions for agent or executor

### Requirement: Legacy jenkins-lib route redirect

Former `/platform/jenkins-lib` bookmarks SHALL redirect to the runtime catalog.

#### Scenario: Redirect jenkins-lib URL

- **WHEN** user navigates to `/platform/jenkins-lib`
- **THEN** the UI MUST redirect to `/platform/runtime`
- **AND** the sidebar MUST NOT highlight a «Jenkins library» nav item (route is not listed in Platform nav)
