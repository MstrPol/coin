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

- **WHEN** enabling team publishes a validated draft to canary for `branching-model`
- **THEN** the UI MUST call Admin API to register the package in PostgreSQL only (artifact bodies + content_ref manifest subset), set state to `canary`, and MUST NOT upload immutable package files to Nexus

#### Scenario: Promote to stable

- **WHEN** health gate passes for pilot projects on canary line and enabling team promotes `branching-model` to stable
- **THEN** the UI MUST upload the immutable package to Nexus, update `content_ref` v2 with `package.url` and digest, and set component state to `published` in one flow

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

### Requirement: Studio entry from platform catalogs

Component Studio SHALL be reachable from platform entity and catalog pages without a dedicated top-level navigation peer.

#### Scenario: Open Studio from branching catalog

- **WHEN** enabling team selects a branching model version in the Platform branching catalog
- **THEN** the UI MUST provide a link to `/studio/branching-model/{name}/{version}`

#### Scenario: Open Studio from build stack

- **WHEN** enabling team selects edit on a gp-content version in Build stacks or GP Build stack tab
- **THEN** the UI MUST link to `/studio/gp-content/{name}/{version}`

#### Scenario: Publisher Studio shortcut

- **WHEN** user with publisher role views the sidebar
- **THEN** the UI MAY show an optional Studio shortcut in the sidebar footer linking to `/studio`

