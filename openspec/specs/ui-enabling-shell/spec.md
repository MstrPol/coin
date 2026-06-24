# ui-enabling-shell Specification

## Purpose

Enabling-team operator console shell: sidebar IA, Fleet vs Platform grouping, legacy route redirects.

## Requirements

### Requirement: Sidebar navigation shell

The coin-ui SHALL use a left sidebar navigation grouped by operator concern instead of a single horizontal top navigation bar.

#### Scenario: View navigation groups

- **WHEN** an authenticated user opens any coin-ui page
- **THEN** the UI MUST show sidebar groups: Overview, Fleet, Golden Paths, Platform, and Admin (admin-only items where applicable)

#### Scenario: Reader sees Platform catalogs

- **WHEN** user has role `reader`
- **THEN** the Platform group MUST include Runtime, Build stacks, Branching models, and Jenkins library entries (read-only)

#### Scenario: Publisher Studio shortcut

- **WHEN** user has role `publisher` or `admin`
- **THEN** the sidebar footer MUST show a Studio shortcut linking to `/studio`

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
