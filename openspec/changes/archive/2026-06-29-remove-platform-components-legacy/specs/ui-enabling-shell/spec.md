## MODIFIED Requirements

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
