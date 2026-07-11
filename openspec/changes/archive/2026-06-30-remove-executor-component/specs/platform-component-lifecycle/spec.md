## MODIFIED Requirements

### Requirement: Two-state component lifecycle

Platform components `gp-content`, `branching-model`, and `agent` SHALL use a two-state lifecycle: `draft` and `published`.

The `canary` status SHALL NOT exist for component versions.

Component type `executor` SHALL NOT be registered, resolved, or validated as a platform component. CI runtime is fully defined by the pinned `agent` version.

#### Scenario: Create draft component version

- **WHEN** enabling team creates a new `gp-content`, `branching-model`, or `agent` version from Platform UI or Jenkins register
- **THEN** coin-api MUST store the version with `status = draft`
- **AND** for `gp-content` and `branching-model` artifact bodies MUST be editable in PostgreSQL until publish
- **AND** for `agent` metadata (`image`, `digest`) MUST be editable while `status = draft`
- **AND** MUST NOT create or require a paired `executor` component version

#### Scenario: Publish draft to stable

- **WHEN** enabling team promotes a validated draft component version from Platform UI
- **THEN** coin-api MUST set `status = published` and make the version immutable
- **AND** for `gp-content` and `branching-model` MUST upload immutable package to Nexus as today
- **AND** for `agent` MUST NOT require Nexus content_ref; `metadata.image` and `metadata.digest` MUST be present and valid before promote
- **AND** CI register flows MUST NOT promote `agent` versions automatically
- **AND** MUST NOT transition through a `canary` component status
- **AND** MUST NOT auto-publish or require `executor/coin-executor@{same version}`

#### Scenario: Delete draft component version

- **WHEN** enabling team deletes a component version with `status = draft`
- **THEN** coin-api MUST remove draft rows and artifact bodies from PostgreSQL
- **AND** MUST NOT modify Nexus

#### Scenario: Published component immutable

- **WHEN** enabling team attempts to edit artifacts or metadata of a `published` component version
- **THEN** coin-api MUST reject the request with HTTP 409 Conflict
