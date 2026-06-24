## MODIFIED Requirements

### Requirement: Branching models navigation entry

The coin-ui SHALL provide branching models under the Platform sidebar group, not as a top-level peer of Fleet and Golden Paths.

#### Scenario: Open branching models from Platform

- **WHEN** enabling team clicks "Branching models" under Platform in the sidebar
- **THEN** the UI MUST navigate to `/platform/branching-models` showing only `branching-model` components

#### Scenario: Legacy URL redirect

- **WHEN** user opens `/branching-models`
- **THEN** the UI MUST redirect to `/platform/branching-models`
