## MODIFIED Requirements

### Requirement: Branching model rule builder

The branching model editor SHALL provide an ordered list of branch rule cards editing schema v2 fields: `name`, `pattern`, `versioning.template`, and `publish`.

The editor SHALL NOT call a branching preview API or evaluate branch/version/publish scenarios in `coin-api`.

#### Scenario: Editor maps to model.yaml

- **WHEN** publisher saves a draft branching model from the editor
- **THEN** the persisted `model.yaml` MUST be valid schema v2 and reflect card order as `branches` list order

#### Scenario: Reorder branch rules

- **WHEN** publisher reorders branch cards in the editor
- **THEN** the YAML `branches` order MUST update (first match wins)

#### Scenario: Editor does not run platform preview

- **WHEN** publisher edits branch rule fields
- **THEN** the UI MUST NOT call `POST /v1/admin/branching-models/preview`
- **AND** MUST NOT show executor-backed scenario results

## REMOVED Requirements

### Requirement: Branch name test with pattern hint

**Reason**: The branch test field depends on the removed branching preview API. Keeping it would require either `coin-api -> coin-executor`, duplicated evaluator logic, or a new shared package, all of which are out of scope for local pilot cleanup.

**Migration**: Use schema/card validation and documentation for authoring. Runtime behavior is evaluated by `coin-executor` during CI.

The editor SHALL provide a test branch name field with debounced preview API feedback for match and errors.

#### Scenario: Test branch on draft edit

- **WHEN** publisher types a branch name in the test field
- **THEN** the UI MUST call the preview API and show matched rule or validation error

### Requirement: Scenario preview panel

**Reason**: The scenario preview panel depends on the removed branching preview API and executor-backed evaluation inside `coin-api`.

**Migration**: Remove scenario preview UI. Keep branching model card/YAML authoring, validation, docs links, and lifecycle controls.

The editor SHALL show preset and custom scenarios including optional publish request toggle with executor-backed results.

#### Scenario: Custom scenario with publish request

- **WHEN** publisher sets custom branch and enables publish request in preview panel
- **THEN** the UI MUST display publish allowed or denied per preview API
