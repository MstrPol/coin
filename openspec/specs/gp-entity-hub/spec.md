# gp-entity-hub Specification

## Purpose
TBD - created by archiving change coin-ui-gp-entity-hub. Update Purpose after archive.
## Requirements
### Requirement: GP hub entity page

The coin-ui SHALL provide a GP hub at `/gp/{name}` as the primary place to manage one Golden Path profile.

#### Scenario: Hub tabs

- **WHEN** enabling team opens a GP hub
- **THEN** the UI MUST offer tabs: Overview, Releases, Policy, Canary, and Build stack

#### Scenario: Policy tab content

- **WHEN** enabling team opens the Policy tab for `go-app`
- **THEN** the UI MUST show the same version policy controls as the former `/catalog` page for that GP

#### Scenario: Canary tab content

- **WHEN** enabling team opens the Canary tab for `go-app`
- **THEN** the UI MUST show the same canary rollout controls as the former `/canary` page for that GP

#### Scenario: Releases tab content

- **WHEN** enabling team opens the Releases tab
- **THEN** the UI MUST list releases for that GP only (published and drafts per existing filters)

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

