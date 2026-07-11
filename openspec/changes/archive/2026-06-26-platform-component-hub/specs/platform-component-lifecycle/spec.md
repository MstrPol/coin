## MODIFIED Requirements

### Requirement: Two-state component lifecycle

Platform components `gp-content`, `branching-model`, and `agent` SHALL use a two-state lifecycle: `draft` and `published`.

The `canary` status SHALL NOT exist for component versions.

`executor` versions SHALL remain derived from agent stack versions at resolve time and SHALL NOT have an independent draft UI lifecycle.

#### Scenario: Create draft component version

- **WHEN** enabling team creates a new `gp-content`, `branching-model`, or `agent` version from Platform UI or Jenkins register
- **THEN** coin-api MUST store the version with `status = draft`
- **AND** for `gp-content` and `branching-model` artifact bodies MUST be editable in PostgreSQL until publish
- **AND** for `agent` metadata (image, digest, architecture) MUST be editable while `status = draft`

#### Scenario: Publish draft to stable

- **WHEN** enabling team or CI promotes a validated draft component version
- **THEN** coin-api MUST set `status = published` and make the version immutable
- **AND** for `gp-content` and `branching-model` MUST upload immutable package to Nexus as today
- **AND** for `agent` MUST NOT require Nexus content_ref; image metadata MUST be sufficient
- **AND** MUST NOT transition through a `canary` component status

#### Scenario: Delete draft component version

- **WHEN** enabling team deletes a component version with `status = draft`
- **THEN** coin-api MUST remove draft rows and artifact bodies from PostgreSQL
- **AND** MUST NOT modify Nexus

#### Scenario: Published component immutable

- **WHEN** enabling team attempts to edit artifacts or metadata of a `published` component version
- **THEN** coin-api MUST reject the request with HTTP 409 Conflict

### Requirement: Component resolve by channel

coin-api resolve SHALL apply component status rules based on resolve channel and GP release status.

#### Scenario: Stable channel requires published pins

- **WHEN** product CI resolves GP on stable channel
- **THEN** all composition pins (`agent`, `gp-content`, `branching-model`) MUST have `status = published`

#### Scenario: Canary channel allows draft component pins

- **WHEN** product CI resolves GP on canary channel
- **AND** the resolved GP release is a draft or is designated on the canary line
- **THEN** `gp-content` and `branching-model` pins MAY have `status = draft` or `published`
- **AND** `agent` pin MUST have `status = published`

#### Scenario: GP draft composition allows draft pins

- **WHEN** publisher creates or updates a GP draft release
- **THEN** coin-api MUST accept `gp-content` and `branching-model` pins with `status = draft` or `published`
- **AND** MUST require `agent` pin with `status = published`

### Requirement: GP promote requires published component pins

Promoting a GP draft to `published` SHALL require every composition pin to reference a `published` component version.

#### Scenario: Promote blocked by draft gp-content

- **WHEN** publisher promotes GP draft with `gp-content/go-app@1.2.0-draft` where that version has `status = draft`
- **THEN** coin-api MUST reject with HTTP 409 Conflict
- **AND** the error MUST list blocking pins with type, name, version, and status

#### Scenario: Promote succeeds with all published pins

- **WHEN** publisher promotes GP draft where agent, gp-content, and branching-model pins are all `published`
- **THEN** coin-api MUST transition GP release to `published`

## REMOVED Requirements

### Requirement: Runtime components published only

**Reason**: Agent stack now uses the same draft → published lifecycle as other platform components; executor remains derived.

**Migration**: Use `POST .../versions/drafts` for Jenkins agent register; promote via admin API or Platform hub UI. Existing `published` agent versions remain valid.
