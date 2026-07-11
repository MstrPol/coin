## MODIFIED Requirements

### Requirement: Runtime publish guidance

The runtime catalog SHALL surface script-first publish path without in-app draft editors.

#### Scenario: Show publish hint

- **WHEN** enabling team views runtime catalog or component detail
- **THEN** the catalog MUST show a link or hint to the publish runbook for registering new agent or executor versions
- **AND** MUST NOT reference Component Studio

#### Scenario: No draft runtime versions in UI

- **WHEN** enabling team views `/platform/runtime`
- **THEN** the UI MUST list only `published` runtime versions for composition selection contexts
- **AND** MUST NOT offer create-draft or edit actions for agent or executor
