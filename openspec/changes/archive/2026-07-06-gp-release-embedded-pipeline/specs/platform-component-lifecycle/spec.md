## MODIFIED Requirements

### Requirement: Two-state component lifecycle

Platform components `branching-model` and `agent` SHALL use a two-state lifecycle: `draft` and `published`. Component type `gp-content` SHALL NOT be registered or published in the platform component registry.

#### Scenario: Create draft component version

- **WHEN** enabling team creates a new `branching-model` or `agent` version from Platform UI or Jenkins register
- **THEN** coin-api MUST store the version with `status = draft`
- **AND** for `branching-model` artifact bodies MUST be editable in PostgreSQL until publish
- **AND** for `agent` metadata (`image`, `digest`) MUST be editable while `status = draft`

#### Scenario: Publish draft to stable

- **WHEN** enabling team promotes a validated draft `branching-model` or `agent` component version from Platform UI
- **THEN** coin-api MUST set `status = published` and make the version immutable

### Requirement: Component resolve by channel

coin-api resolve SHALL apply component status rules based on resolve channel and GP release status for external composition pins.

#### Scenario: Stable channel requires published external pins

- **WHEN** product CI resolves GP on stable channel
- **THEN** agent and branching-model composition pins MUST have `status = published`

#### Scenario: Canary channel allows draft branching-model pin

- **WHEN** product CI resolves GP on canary channel
- **AND** the resolved GP release is a draft or is designated on the canary line
- **THEN** branching-model pin MAY have `status = draft` or `published`
- **AND** agent pin MUST have `status = published`

#### Scenario: GP draft composition allows draft branching pin

- **WHEN** publisher creates or updates a GP draft release
- **THEN** coin-api MUST accept branching-model pin with `status = draft` or `published`
- **AND** MUST require `agent` pin with `status = published`

### Requirement: GP promote requires published component pins

Promoting a GP draft to `published` SHALL require every external composition pin to reference a `published` component version and embedded pipeline to be valid.

#### Scenario: Promote blocked by draft branching-model pin

- **WHEN** publisher promotes GP draft with branching-model pin in `draft` status
- **THEN** coin-api MUST reject with HTTP 409 Conflict

#### Scenario: Promote succeeds with published external pins and valid pipeline

- **WHEN** publisher promotes GP draft where agent and branching-model pins are `published` and embedded pipeline is valid
- **THEN** coin-api MUST transition GP release to `published`

### Requirement: Admin API delete component draft

coin-api SHALL expose `DELETE /v1/admin/components/{type}/{name}/versions/{version}` for platform component types with draft lifecycle: `agent` and `branching-model` only.

#### Scenario: Delete branching-model draft succeeds

- **WHEN** publisher deletes a `branching-model` draft version
- **THEN** coin-api MUST return HTTP 204 No Content and cascade-delete artifact bodies

## REMOVED Requirements

### Requirement: Two-state component lifecycle (gp-content clause)

**Reason**: gp-content component type removed; pipeline lifecycle is GP release lifecycle.

**Migration**: Edit and promote embedded pipeline on GP release detail.
