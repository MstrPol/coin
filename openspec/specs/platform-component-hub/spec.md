# platform-component-hub Specification

## Purpose
TBD - created by archiving change platform-component-hub. Update Purpose after archive.
## Requirements
### Requirement: Platform component hub entity page

The coin-ui SHALL provide a Platform component hub at `/platform/{family}/{name}` as the primary place to manage one platform component profile, where `{family}` is `runtime`, `build-stacks`, or `branching-models`.

#### Scenario: Hub tabs for platform families

- **WHEN** enabling team opens a platform component hub for any supported family
- **THEN** the UI MUST offer tabs: Overview and Releases
- **AND** MUST NOT offer GP-only tabs (Policy, Canary)

#### Scenario: Overview tab empty state

- **WHEN** enabling team opens Overview for a profile with no versions
- **THEN** the UI MUST show profile metadata (name, family label) and an empty-state CTA to create the first draft

#### Scenario: Releases tab lists profile versions only

- **WHEN** enabling team opens the Releases tab for `build-stacks/go-app`
- **THEN** the UI MUST list versions for `gp-content/go-app` only, grouped by status (`draft`, `published`)

### Requirement: Hub draft-only primary action

The platform component hub SHALL expose «New draft» as the single primary publisher action for new version work.

#### Scenario: Hub action without direct publish

- **WHEN** publisher views a platform component hub
- **THEN** the UI MUST show «New draft» as the primary action
- **AND** MUST NOT show «New release» or «Publish» as a primary hub action

#### Scenario: Welcome after profile create

- **WHEN** publisher lands on hub after creating a profile (`?welcome=1`)
- **THEN** the UI MUST prompt to create the first draft

### Requirement: Platform family catalog shows profiles

Each platform family catalog (`/platform/runtime`, `/platform/build-stacks`, `/platform/branching-models`) SHALL list component **profiles** (names), not a flat list of all versions across profiles.

#### Scenario: Catalog row navigates to hub

- **WHEN** enabling team opens `/platform/build-stacks` and selects profile `go-app`
- **THEN** the UI MUST navigate to `/platform/build-stacks/go-app`

#### Scenario: Catalog primary action for new profile

- **WHEN** publisher with publisher role views a platform family catalog
- **THEN** the UI MUST show «New profile» as the catalog-level create action
- **AND** MUST NOT use the label «Create draft» at catalog level

### Requirement: Platform hub URL tabs

Platform component hub tab state SHALL be reflected in the URL path for bookmarking.

#### Scenario: Bookmark releases tab

- **WHEN** user opens `/platform/runtime/coin-agent/releases`
- **THEN** the UI MUST show the Releases tab for agent profile `coin-agent`

### Requirement: Release detail under platform hub

Platform component release detail SHALL live under the platform hub URL hierarchy.

#### Scenario: Agent release detail shows derived executor

- **WHEN** enabling team opens release detail for agent `coin-agent` version `1.0.0`
- **THEN** the UI MUST show agent metadata (`image`, `digest`)
- **AND** MUST show read-only derived pin `executor/coin-executor@1.0.0`
- **AND** MUST NOT list executor as a separate release in the Releases tab
- **AND** MUST NOT show GOARCH or architecture fields

#### Scenario: Draft release detail actions

- **WHEN** publisher opens release detail for a draft platform component version
- **THEN** the UI MUST offer promote and delete draft actions appropriate to the component type
- **AND** for `agent` drafts MUST wire delete draft to `DELETE /v1/admin/components/agent/{name}/versions/{version}`
- **AND** MUST NOT offer delete for published releases

### Requirement: Legacy flat URL redirects

Former flat platform version URLs SHALL redirect into the hub hierarchy.

#### Scenario: Redirect build stack release detail

- **WHEN** user opens `/platform/build-stacks/go-app/1.0.0`
- **THEN** the UI MUST redirect to `/platform/build-stacks/go-app/releases/1.0.0`

#### Scenario: Redirect runtime component detail

- **WHEN** user opens `/components/agent/coin-agent`
- **THEN** the UI MUST redirect to `/platform/runtime/coin-agent`

