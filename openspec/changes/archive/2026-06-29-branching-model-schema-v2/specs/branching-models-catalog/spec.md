## ADDED Requirements

### Requirement: Branching model rule builder

The branching model editor SHALL provide an ordered list of branch rule cards editing schema v2 fields: `name`, `pattern`, `versioning.template`, and `publish`.

#### Scenario: Editor maps to model.yaml

- **WHEN** publisher saves a draft branching model from the editor
- **THEN** the persisted `model.yaml` MUST be valid schema v2 and reflect card order as `branches` list order

#### Scenario: Reorder branch rules

- **WHEN** publisher reorders branch cards in the editor
- **THEN** the YAML `branches` order MUST update (first match wins)

### Requirement: Branch name test with pattern hint

The editor SHALL provide a test branch name field with debounced preview API feedback for match and errors.

#### Scenario: Test branch on draft edit

- **WHEN** publisher types a branch name in the test field
- **THEN** the UI MUST call the preview API and show matched rule or validation error

### Requirement: Scenario preview panel

The editor SHALL show preset and custom scenarios including optional publish request toggle with executor-backed results.

#### Scenario: Custom scenario with publish request

- **WHEN** publisher sets custom branch and enables publish request in preview panel
- **THEN** the UI MUST display publish allowed or denied per preview API

### Requirement: Branching model authoring documentation

The branching models catalog and editor SHALL link operators to `docs/how-to/branching-models.md` and per-model README under `coin-branching-models/models/{name}/`.

#### Scenario: How-to link from editor

- **WHEN** publisher opens branching model draft editor
- **THEN** the UI MUST provide a link or reference to the authoring how-to documentation
