# component-studio Specification

## Purpose

Component Studio as primary UI-first authoring path for platform components.

## Requirements
### Requirement: Component Studio as primary authoring path

The coin-ui SHALL provide Component Studio as the primary path for enabling team to create and publish platform components without git or shell scripts.

#### Scenario: Create draft component

- **WHEN** enabling team creates a new component version in Component Studio
- **THEN** the UI MUST save it as `draft` in PostgreSQL and MUST NOT require a git commit or publish script

#### Scenario: Validate before publish

- **WHEN** enabling team clicks Validate
- **THEN** the UI MUST run server-side schema validation and show errors before allowing publish to canary

#### Scenario: Publish to canary

- **WHEN** enabling team publishes a validated draft to canary
- **THEN** the UI MUST call Admin API to upload package to Nexus, register version, and set state to `canary`

#### Scenario: Promote to stable

- **WHEN** health gate passes for pilot projects on canary line
- **THEN** the UI MUST offer Promote to stable updating component state and catalog pointers in one flow

### Requirement: Type-aware editors

Component Studio SHALL render type-specific forms instead of raw JSON for structured component types.

#### Scenario: branching-model editor (GCP-1 green field)

- **WHEN** editing `branching-model` component type
- **THEN** the UI MUST provide form fields for `model.yaml` structure (first green field type in GCP-1)

#### Scenario: gp-content editor (GCP-3)

- **WHEN** editing `gp-content` after GCP-3
- **THEN** the UI MUST provide fields for content.yaml and Containerfile editing

### Requirement: Pilot project selection

The UI SHALL allow selecting pilot projects for canary testing before stable promote.

#### Scenario: Assign pilots

- **WHEN** enabling team configures a canary GP release
- **THEN** the UI MUST allow picking projects with `canary_mode=canary` for smoke/E2E validation

