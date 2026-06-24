# platform-build-stacks Specification

## Purpose

Platform catalog and GP detail integration for gp-content build stacks.
## Requirements
### Requirement: Build stacks catalog

The coin-ui SHALL provide a Platform → Build stacks catalog for all `gp-content` components.

#### Scenario: List gp-content stacks

- **WHEN** enabling team opens `/platform/build-stacks`
- **THEN** the UI MUST list `gp-content` components with GP profile name alignment (e.g. `go-app`), versions, and lifecycle status

#### Scenario: Open stack in Studio

- **WHEN** enabling team selects a gp-content version
- **THEN** the UI MUST link to Component Studio at `/studio/gp-content/{name}/{version}` when publisher role is present

### Requirement: GP detail Build stack tab

The GP release detail page SHALL include a Build stack tab as the primary path to gp-content for that profile.

#### Scenario: View build stack from GP hub

- **WHEN** enabling team opens the Build stack tab on GP hub for profile `go-app`
- **THEN** the UI MUST show gp-content versions for that GP name (primary path)

#### Scenario: Release detail defers to hub

- **WHEN** enabling team views release detail for a GP release
- **THEN** the UI MAY link to the GP hub Build stack tab rather than duplicating full build stack management on release detail

