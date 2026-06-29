# ui-enabling-shell Specification

## Purpose

Enabling-team operator console shell: sidebar IA, Fleet vs Platform grouping, legacy route redirects.
## Requirements
### Requirement: Sidebar navigation shell

The coin-ui SHALL use a left sidebar navigation grouped by operator concern instead of a single horizontal top navigation bar.

#### Scenario: View navigation groups

- **WHEN** an authenticated user opens any coin-ui page
- **THEN** the UI MUST show sidebar groups: Overview, Fleet, Golden Paths, Platform, and Admin (admin-only items where applicable)

#### Scenario: Golden Paths nav entries

- **WHEN** user views the Golden Paths group in the sidebar
- **THEN** the UI MUST include GP Profiles (`/gp`) and Resolve (`/resolve`) only — not separate Releases, GP Policy, or Canary top-level items

### Requirement: Legacy route redirects

The coin-ui SHALL preserve bookmarks from pre-IA routes via redirects to the current Platform family catalogs.

#### Scenario: Redirect branching models

- **WHEN** user navigates to `/branching-models`
- **THEN** the UI MUST redirect to `/platform/branching-models` without losing catalog functionality

#### Scenario: Redirect components list

- **WHEN** user navigates to `/components`
- **THEN** the UI MUST redirect to `/platform/runtime` (Platform default catalog)

#### Scenario: Redirect legacy platform components aggregate

- **WHEN** user navigates to `/platform/components`
- **THEN** the UI MUST redirect to `/platform/runtime`
- **AND** MUST NOT render an aggregate all-types components page

#### Scenario: Redirect legacy component detail by type

- **WHEN** user navigates to `/components/:type/:name` for a component type with a Platform family mapping (`agent`, `gp-content`, `branching-model`)
- **THEN** the UI MUST redirect to the corresponding family hub (`/platform/runtime/:name`, `/platform/build-stacks/:name`, or `/platform/branching-models/:name`)

### Requirement: Full-width main content

The shell SHALL allow catalog and entity pages to use available viewport width.

#### Scenario: Remove narrow shell constraint

- **WHEN** user views a catalog table under the new shell
- **THEN** main content MUST NOT be constrained to `max-w-5xl` at the layout level

### Requirement: Platform navigation without Jenkins library

The Platform sidebar group SHALL list runtime, build stacks, and branching models only — not Jenkins library management.

#### Scenario: Platform nav items

- **WHEN** user views the Platform group in the sidebar
- **THEN** the UI MUST include Runtime, Build stacks, and Branching models
- **AND** MUST NOT include a Jenkins library entry

### Requirement: Legacy platform settings redirect

The coin-ui SHALL redirect former platform settings bookmarks.

#### Scenario: Redirect platform settings URL

- **WHEN** user navigates to `/platform-settings`
- **THEN** the UI MUST redirect to `/audit`

### Requirement: Admin navigation without platform settings

The Admin sidebar group SHALL NOT include a Platform settings entry.

#### Scenario: Admin nav items

- **WHEN** admin views the Admin group in the sidebar
- **THEN** the UI MUST include Audit (`/audit`)
- **AND** MUST NOT include Platform settings or Nexus configuration

