## MODIFIED Requirements

### Requirement: Model schema

Each branching model SHALL conform to `branching-model.schema.json` **schemaVersion 2** with an ordered `branches` list. Each branch rule MUST define `name`, `pattern` (RE2), `versioning.template` (mini-DSL), and `publish` (boolean eligibility).

#### Scenario: Invalid model rejected

- **WHEN** model.yaml fails schema v2 validation or template placeholder rules
- **THEN** validate-package MUST reject before registry write

#### Scenario: Schema v1 rejected

- **WHEN** model.yaml has `schemaVersion: 1` or v1-only fields (`trunk`, `branchTypes`, `publish.when`)
- **THEN** coin-api MUST reject the draft

#### Scenario: Main branch rule present

- **WHEN** a valid v2 model is validated
- **THEN** it MUST include a branch rule matching `main` or `master` (e.g. `pattern: ^main$|^master$`)
