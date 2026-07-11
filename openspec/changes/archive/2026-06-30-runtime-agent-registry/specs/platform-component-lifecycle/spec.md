## MODIFIED Requirements

### Requirement: Two-state component lifecycle

Platform components `gp-content`, `branching-model`, and `agent` SHALL use a two-state lifecycle: `draft` and `published`.

The `canary` status SHALL NOT exist for component versions.

`executor` versions SHALL remain derived from agent stack versions at resolve time as `executor/coin-executor@{same version as agent pin}` for any registered `agent` profile, and SHALL NOT have an independent draft UI lifecycle.

#### Scenario: Executor derived for any agent profile

- **WHEN** GP composition pins `agent/{profile}@{version}` for any registered profile name
- **THEN** coin-api MUST derive `executor/coin-executor@{version}` at resolve time
- **AND** MUST NOT reject derive solely because profile name is not `coin-agent`

#### Scenario: Create draft component version

- **WHEN** enabling team creates a new `gp-content`, `branching-model`, or `agent` version from Platform UI or Jenkins register
- **THEN** coin-api MUST store the version with `status = draft`
- **AND** for `gp-content` and `branching-model` artifact bodies MUST be editable in PostgreSQL until publish
- **AND** for `agent` metadata (`image`, `digest`) MUST be editable while `status = draft`

#### Scenario: Publish draft to stable

- **WHEN** enabling team promotes a validated draft component version from Platform UI
- **THEN** coin-api MUST set `status = published` and make the version immutable
- **AND** for `gp-content` and `branching-model` MUST upload immutable package to Nexus as today
- **AND** for `agent` MUST NOT require Nexus content_ref; `metadata.image` and `metadata.digest` MUST be present and valid before promote
- **AND** CI register flows MUST NOT promote `agent` versions automatically
- **AND** MUST NOT transition through a `canary` component status

#### Scenario: Delete draft component version

- **WHEN** enabling team deletes a component version with `status = draft`
- **THEN** coin-api MUST remove draft rows and artifact bodies from PostgreSQL
- **AND** MUST NOT modify Nexus

#### Scenario: Published component immutable

- **WHEN** enabling team attempts to edit artifacts or metadata of a `published` component version
- **THEN** coin-api MUST reject the request with HTTP 409 Conflict
