## ADDED Requirements

### Requirement: Containerfiles catalog panel

coin-ui SHALL provide a Containerfiles catalog editor for GP drafts: list entries (`id`, `kind`, `path`), create/edit/delete managed and project entries, without requiring inline body in the resolved preview path.

#### Scenario: Add project containerfile entry

- **WHEN** user adds a project containerfile with id and path
- **THEN** draft MUST persist a catalog entry `kind: project` with that path

### Requirement: Task graph editor

coin-ui SHALL edit `pipeline.tasks` including `id`, `runAfter`, and typed steps (`coin`, `containerfile`, `sh`). Graph or list UI MUST preserve semantic ids and step refs to catalog / destinations.

#### Scenario: Link publish to build task

- **WHEN** user configures a publish step with `buildTaskId` and `destinationRefs`
- **THEN** draft MUST store those fields on the publish step
- **AND** preview MUST surface them in resolved tasks

### Requirement: Entity hub layout for v4 pipeline

GP entity hub layout SHALL order sections Composition → Pipeline → Containerfiles → Parameters for schemaVersion 4 drafts.

#### Scenario: Navigation order

- **WHEN** user opens a v4 GP draft in entity hub
- **THEN** UI MUST expose Pipeline and Containerfiles sections in that IA order
