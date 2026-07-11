## ADDED Requirements

### Requirement: Draft creation without platform lib prerequisite

Creating a GP draft SHALL NOT require a configured platform lib pin or validation against lib registry entries.

#### Scenario: New draft without platform runtime

- **WHEN** publisher creates a new GP draft with valid agent, gp-content, and branching-model pins
- **AND** platform settings contain no runtime/lib configuration
- **THEN** coin-api MUST accept the draft
- **AND** the UI MUST NOT block draft submission due to missing lib pin
