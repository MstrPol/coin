# gp-entity-hub Specification

## Purpose
TBD - created by archiving change coin-ui-gp-entity-hub. Update Purpose after archive.
## Requirements
### Requirement: GP hub entity page

The coin-ui SHALL provide a GP hub at `/gp/{name}` as the primary place to manage one Golden Path profile.

#### Scenario: Hub tabs

- **WHEN** enabling team opens a GP hub
- **THEN** the UI MUST offer tabs: Overview, Releases, Policy, and Canary
- **AND** MUST NOT offer a Build stack tab on the profile hub (no profile ↔ gp-content relationship)

#### Scenario: Policy tab content

- **WHEN** enabling team opens the Policy tab for `go-app`
- **THEN** the UI MUST show the same version policy controls as the former `/catalog` page for that GP

#### Scenario: Canary tab content

- **WHEN** enabling team opens the Canary tab for `go-app`
- **THEN** the UI MUST show the same canary rollout controls as the former `/canary` page for that GP

#### Scenario: Releases tab content

- **WHEN** enabling team opens the Releases tab
- **THEN** the UI MUST list releases for that GP only (published and drafts per existing filters)

#### Scenario: Release detail shows version composition

- **WHEN** enabling team opens release detail for GP `xxx` version `1.0.0`
- **THEN** the UI MUST show the composition table for **that version** (agent, gp-content, branching-model pins)
- **AND** agent pin MUST link to `/platform/runtime/{agentName}/releases/{version}`
- **AND** gp-content pin MUST link to `/platform/build-stacks/{name}/releases/{version}` or edit route for draft
- **AND** branching-model pin MUST link to `/platform/branching-models/{name}/releases/{version}` or edit route for draft
- **AND** MUST NOT link to flat catalog URLs or Component Studio

#### Scenario: Draft release detail actions

- **WHEN** publisher opens release detail for a draft
- **THEN** the UI MUST offer promote and delete draft actions
- **AND** MUST NOT offer delete for published releases

### Requirement: Hub draft-only primary action

The GP hub SHALL expose a single primary publisher action for new composition work: create draft.

#### Scenario: Hub actions without direct publish

- **WHEN** publisher views GP hub for a profile with no releases
- **THEN** the UI MUST show «New draft» as the primary action and MUST NOT show «New release»

#### Scenario: Welcome after profile create

- **WHEN** publisher lands on hub after creating a profile (`?welcome=1`)
- **THEN** the UI MUST prompt to create the first draft (not direct publish)

### Requirement: Overview without profile slots

The GP hub Overview tab SHALL NOT display a composition slots table sourced from the profile entity.

#### Scenario: Empty profile overview

- **WHEN** enabling team opens Overview for a profile with no drafts or releases
- **THEN** the UI MUST show profile description (if any) and an empty-state CTA to create a draft

### Requirement: GP hub URL tabs

GP hub tab state SHALL be reflected in the URL path for bookmarking.

#### Scenario: Bookmark policy tab

- **WHEN** user opens `/gp/go-app/policy`
- **THEN** the UI MUST show the Policy tab for `go-app`

### Requirement: Legacy policy and canary redirects

Former global policy and canary pages SHALL redirect into the GP hub.

#### Scenario: Redirect catalog

- **WHEN** user opens `/catalog` with `name=go-app`
- **THEN** the UI MUST redirect to `/gp/go-app/policy`

#### Scenario: Redirect canary

- **WHEN** user opens `/canary` with `name=go-app`
- **THEN** the UI MUST redirect to `/gp/go-app/canary`

### Requirement: Release detail under GP hub

GP release detail SHALL live under the GP hub URL hierarchy.

#### Scenario: Redirect release detail

- **WHEN** user opens `/releases/go-app/1.0.0`
- **THEN** the UI MUST redirect to `/gp/go-app/releases/1.0.0`

