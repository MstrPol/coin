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

The coin-ui SHALL preserve bookmarks from pre-IA routes via redirects.

#### Scenario: Redirect branching models

- **WHEN** user navigates to `/branching-models`
- **THEN** the UI MUST redirect to `/platform/branching-models` without losing catalog functionality

#### Scenario: Redirect components list

- **WHEN** user navigates to `/components`
- **THEN** the UI MUST redirect to `/platform/components` (legacy aggregate view)

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

### Requirement: Platform settings without runtime lib pin

The Platform settings admin page SHALL configure Nexus integration only and MUST NOT expose lib pin editing.

The page MUST render successfully when the platform settings API response omits a `runtime` field (coin-lib is outside control plane).

#### Scenario: Settings form fields

- **WHEN** admin opens `/platform-settings`
- **THEN** the UI MUST show Nexus configuration fields
- **AND** MUST NOT show platform lib pin fields or runtime section

#### Scenario: Settings without runtime in API response

- **WHEN** admin opens `/platform-settings`
- **AND** `GET /v1/admin/platform/settings` returns only `nexusMavenBase`, `nexusCredentialsId`, and `updatedAt`
- **THEN** the UI MUST render the page without JavaScript errors
- **AND** MUST NOT access or display `runtime.lib`

