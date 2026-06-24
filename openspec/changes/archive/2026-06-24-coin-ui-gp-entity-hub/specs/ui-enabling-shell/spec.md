## MODIFIED Requirements

### Requirement: Sidebar navigation shell

The coin-ui SHALL use a left sidebar navigation grouped by operator concern instead of a single horizontal top navigation bar.

#### Scenario: View navigation groups

- **WHEN** an authenticated user opens any coin-ui page
- **THEN** the UI MUST show sidebar groups: Overview, Fleet, Golden Paths, Platform, and Admin (admin-only items where applicable)

#### Scenario: Golden Paths nav entries

- **WHEN** user views the Golden Paths group in the sidebar
- **THEN** the UI MUST include GP Profiles (`/gp`) and Resolve (`/resolve`) only — not separate Releases, GP Policy, or Canary top-level items
