# platform-component-lifecycle Specification

## Purpose

Platform-native two-state lifecycle for gp-content and branching-model; GP promote gate and canary resolve rules.

## Requirements

### Requirement: Two-state component lifecycle

Platform components `gp-content` and `branching-model` SHALL use a two-state lifecycle: `draft` and `published`.

The `canary` status SHALL NOT exist for component versions after this change.

#### Scenario: Create draft component version

- **WHEN** enabling team creates a new `gp-content` or `branching-model` version from Platform UI
- **THEN** coin-api MUST store the version with `status = draft`
- **AND** artifact bodies MUST be editable in PostgreSQL until publish

#### Scenario: Publish draft to stable

- **WHEN** enabling team publishes a validated draft component version
- **THEN** coin-api MUST upload immutable package to Nexus, set `status = published`, and make the version immutable
- **AND** MUST NOT transition through a `canary` component status

#### Scenario: Delete draft component version

- **WHEN** enabling team deletes a component version with `status = draft`
- **THEN** coin-api MUST remove draft rows and artifact bodies from PostgreSQL
- **AND** MUST NOT modify Nexus

#### Scenario: Published component immutable

- **WHEN** enabling team attempts to edit artifacts of a `published` component version
- **THEN** coin-api MUST reject the request with HTTP 409 Conflict

### Requirement: Runtime components published only

Platform runtime components (`agent`, `executor`) SHALL only exist in `published` status for registry and GP composition purposes.

The coin-ui MUST NOT provide draft create or edit flows for runtime components.

#### Scenario: Agent pin in GP draft

- **WHEN** publisher selects agent stack version in GP draft composition
- **THEN** the UI MUST offer only `published` agent versions
- **AND** coin-api MUST reject GP draft composition if agent pin is not `published`

#### Scenario: Runtime publish path

- **WHEN** enabling team publishes a new agent or executor version
- **THEN** the primary path MUST remain script-first registration (publish runbook)
- **AND** coin-ui MUST show publish guidance on runtime detail without an in-app editor

### Requirement: Component resolve by channel

coin-api resolve SHALL apply component status rules based on resolve channel and GP release status.

#### Scenario: Stable channel requires published pins

- **WHEN** product CI resolves GP on stable channel
- **THEN** all composition pins (`agent`, `gp-content`, `branching-model`) MUST have `status = published`

#### Scenario: Canary channel allows draft component pins

- **WHEN** product CI resolves GP on canary channel
- **AND** the resolved GP release is a draft or is designated on the canary line
- **THEN** `gp-content` and `branching-model` pins MAY have `status = draft` or `published`
- **AND** `agent` pin MUST still have `status = published`

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

### Requirement: Draft pin instability warning

The coin-ui SHALL warn when a composition or canary line uses draft component pins that may change.

#### Scenario: Draft pin badge in GP composition

- **WHEN** publisher views GP draft composition containing a pin with `status = draft`
- **THEN** the UI MUST display the pin status badge `draft`
- **AND** MUST show warning text that the pin may change before stable publish

#### Scenario: Canary line with draft GP warning

- **WHEN** enabling team assigns a GP draft containing draft component pins to the canary line
- **THEN** the UI MUST show a confirmation warning that pilot projects may receive unstable manifests
- **AND** MUST NOT require locking draft component versions
