## MODIFIED Requirements

### Requirement: Branching model authoring documentation

The branching models catalog and editor SHALL link operators to `docs/how-to/branching-models.md` as the sole human authoring guide. The UI MUST NOT require or link to a separate git catalog under `coin-branching-models/`.

#### Scenario: How-to link from editor

- **WHEN** publisher opens branching model draft editor
- **THEN** the UI MUST provide a link or reference to the authoring how-to documentation
- **AND** MUST NOT depend on per-model README paths under `coin-branching-models/models/{name}/`
