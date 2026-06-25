# gp-profile-catalog Specification

## Purpose
TBD - created by archiving change coin-ui-gp-entity-hub. Update Purpose after archive.

## Requirements

### Requirement: GP profile catalog page

The coin-ui SHALL provide a Golden Paths catalog listing GP **profiles**, not individual release versions.

#### Scenario: View GP profiles

- **WHEN** enabling team opens `/gp`
- **THEN** the UI MUST show one row per GP profile with `name`, optional `description` summary, latest stable, latest canary when available, and release counts — and MUST NOT show a Slots column

#### Scenario: Open GP hub from catalog

- **WHEN** enabling team selects a GP profile from the catalog
- **THEN** the UI MUST navigate to `/gp/{name}` (GP hub Overview tab)

#### Scenario: Create new GP profile

- **WHEN** enabling team starts «New profile» from the catalog
- **THEN** the UI MUST route to a form collecting only `name` and optional `description` without component version pickers

### Requirement: Legacy releases list redirect

The flat global releases list SHALL redirect to the GP profile catalog.

#### Scenario: Redirect old releases URL

- **WHEN** user navigates to `/releases`
- **THEN** the UI MUST redirect to `/gp`
