## ADDED Requirements

### Requirement: Studio entry from platform catalogs

Component Studio SHALL be reachable from platform entity and catalog pages without a dedicated top-level navigation peer.

#### Scenario: Open Studio from branching catalog

- **WHEN** enabling team selects a branching model version in the Platform branching catalog
- **THEN** the UI MUST provide a link to `/studio/branching-model/{name}/{version}`

#### Scenario: Open Studio from build stack

- **WHEN** enabling team selects edit on a gp-content version in Build stacks or GP Build stack tab
- **THEN** the UI MUST link to `/studio/gp-content/{name}/{version}`

#### Scenario: Publisher Studio shortcut

- **WHEN** user with publisher role views the sidebar
- **THEN** the UI MAY show an optional Studio shortcut in the sidebar footer linking to `/studio`
