# platform-runtime-catalog Specification

## Purpose

Platform catalog for runtime components (agent, executor) in coin-ui.
## Requirements
### Requirement: Runtime catalog page

The coin-ui SHALL provide a Platform → Runtime catalog listing **agent stack profiles** only.

The page MUST NOT display a platform lib pin banner, «Current platform lib pin» heading, or `lib` / `executor` component rows.

#### Scenario: No lib pin banner on runtime page

- **WHEN** enabling team opens `/platform/runtime`
- **THEN** the UI MUST NOT show text «Current platform lib pin» or any lib version pin block

#### Scenario: List agent stack profiles

- **WHEN** enabling team opens `/platform/runtime`
- **THEN** the UI MUST show agent component profiles (e.g. `coin-agent`, `coin-agent-arm`) with version summary per profile
- **AND** MUST NOT show type `executor` or `lib` components

#### Scenario: Open agent stack hub from catalog

- **WHEN** enabling team selects an agent profile row
- **THEN** the UI MUST navigate to `/platform/runtime/{name}`

### Requirement: Runtime publish guidance

The runtime catalog and hub SHALL support draft registration from CI and manual catch-up, with promote as a publisher-only Platform action.

#### Scenario: Show publish runbook hint

- **WHEN** enabling team views runtime catalog or agent hub
- **THEN** the UI MUST show a link or hint to the agent publish runbook for CI image push and API draft register
- **AND** MUST state that promote happens in Platform UI after review
- **AND** MUST NOT reference Component Studio

#### Scenario: New agent profile from catalog

- **WHEN** publisher clicks «New profile» on `/platform/runtime`
- **THEN** the UI MUST navigate to `/platform/runtime/new` to create an agent component profile
- **AND** after create MUST land on hub with welcome CTA for first draft

#### Scenario: Draft agent versions visible in hub

- **WHEN** enabling team opens agent stack hub Releases tab
- **THEN** the UI MUST list both `draft` and `published` versions for that profile
- **AND** MUST offer «New draft» on the hub for additional versions

### Requirement: Legacy jenkins-lib route redirect

Former `/platform/jenkins-lib` bookmarks SHALL redirect to the runtime catalog.

#### Scenario: Redirect jenkins-lib URL

- **WHEN** user navigates to `/platform/jenkins-lib`
- **THEN** the UI MUST redirect to `/platform/runtime`
- **AND** the sidebar MUST NOT highlight a «Jenkins library» nav item (route is not listed in Platform nav)

### Requirement: Manual catch-up draft form

When creating an agent draft from Platform UI (image already in registry), the form SHALL require image ref and digest only.

The form MUST derive and display the version as a read-only preview parsed from the image tag.

The form MUST NOT include an editable Version field or GOARCH / architecture fields.

#### Scenario: Create catch-up draft

- **WHEN** publisher opens «New draft» for an agent profile without CI register
- **THEN** the UI MUST require Image ref and Digest before submit
- **AND** MUST show read-only derived version from the image tag (e.g. `1.2.0`)
- **AND** MUST NOT show editable Version or GOARCH inputs

#### Scenario: Disable submit when tag unparseable

- **WHEN** publisher enters an image ref without a valid tag
- **THEN** the UI MUST disable submit
- **AND** MUST show validation hint on Image ref

#### Scenario: CI-registered draft release detail

- **WHEN** publisher opens release detail for agent draft created by CI
- **THEN** the UI MUST display image and digest read-only
- **AND** MUST offer Promote action for publisher role
- **AND** MUST NOT offer GOARCH field

### Requirement: Delete agent draft from runtime hub

The runtime agent hub SHALL allow publishers to delete draft versions from the Releases tab and release detail page.

#### Scenario: Delete draft from releases list

- **WHEN** publisher views `/platform/runtime/{profile}/releases`
- **AND** a version row has `status = draft`
- **THEN** the UI MUST offer a «Delete» or «Delete draft» action for that row
- **AND** MUST NOT offer delete for `published` rows

#### Scenario: Delete draft from release detail

- **WHEN** publisher opens release detail for an agent draft at `/platform/runtime/{profile}/releases/{version}`
- **THEN** the UI MUST offer «Delete draft» alongside Promote
- **AND** MUST require confirmation before calling the Admin API
- **AND** after successful delete MUST navigate back to the Releases tab

#### Scenario: Non-publisher cannot delete

- **WHEN** user without publisher role views agent draft releases
- **THEN** the UI MUST NOT show delete actions

