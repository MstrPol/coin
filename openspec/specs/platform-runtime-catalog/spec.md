# platform-runtime-catalog Specification

## Purpose

Platform catalog for runtime components (agent, executor) in coin-ui.

## Requirements

### Requirement: Runtime catalog page

The coin-ui SHALL provide a Platform → Runtime catalog listing `agent` and `executor` components.

#### Scenario: List runtime components

- **WHEN** enabling team opens `/platform/runtime`
- **THEN** the UI MUST show only components with type `agent` or `executor`, with name, version lines, and lifecycle status

#### Scenario: Open runtime component detail

- **WHEN** enabling team selects a runtime component row
- **THEN** the UI MUST navigate to the existing component detail route for that type and name

### Requirement: Runtime publish guidance

The runtime catalog SHALL surface publish path without Studio for script-first types.

#### Scenario: Show publish hint

- **WHEN** a runtime component has no draft in Studio
- **THEN** the catalog MUST show a link or hint to the publish runbook (external docs or inline note)
