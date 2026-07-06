# platform-component-lifecycle Specification

## Purpose

Platform-native two-state lifecycle for `branching-model` and `agent`; GP promote gate and canary resolve rules. Embedded pipeline lifecycle is GP release lifecycle.

## Requirements

### Requirement: Two-state component lifecycle

Platform components `branching-model` and `agent` SHALL use a two-state lifecycle: `draft` and `published`. Component type `gp-content` SHALL NOT be registered or published in the platform component registry.

The `canary` status SHALL NOT exist for component versions.

Component type `executor` SHALL NOT be registered, resolved, or validated as a platform component. CI runtime is fully defined by the pinned `agent` version.

#### Scenario: Create draft component version

- **WHEN** enabling team creates a new `branching-model` or `agent` version from Platform UI or Jenkins register
- **THEN** coin-api MUST store the version with `status = draft`
- **AND** for `branching-model` artifact bodies MUST be editable in PostgreSQL until publish
- **AND** for `agent` metadata (`image`, `digest`) MUST be editable while `status = draft`
- **AND** MUST NOT create or require a paired `executor` component version

#### Scenario: Publish draft to stable

- **WHEN** enabling team promotes a validated draft `branching-model` or `agent` component version from Platform UI
- **THEN** coin-api MUST set `status = published` and make the version immutable
- **AND** for `branching-model` MUST upload immutable package to Nexus as today
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
- **AND** the error MUST list blocking pins with type, name, version, and status

#### Scenario: Promote succeeds with published external pins and valid pipeline

- **WHEN** publisher promotes GP draft where agent and branching-model pins are `published` and embedded pipeline is valid
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

### Requirement: Admin API delete component draft

coin-api SHALL expose `DELETE /v1/admin/components/{type}/{name}/versions/{version}` for platform component types with draft lifecycle: `agent` and `branching-model` only.

The endpoint MUST NOT delete versions with `status = published`.

#### Scenario: Delete agent draft succeeds

- **WHEN** publisher calls `DELETE /v1/admin/components/agent/coin-agent/versions/0.1.0-draft`
- **AND** the version has `status = draft`
- **THEN** coin-api MUST return HTTP 204 No Content
- **AND** MUST remove the `component_versions` row
- **AND** MUST cascade-delete any `component_artifact_bodies` rows for that version
- **AND** MUST write audit log action `delete_component_draft`
- **AND** MUST NOT modify Nexus

#### Scenario: Reject delete published version

- **WHEN** publisher attempts to delete `agent/coin-agent@1.0.0` with `status = published`
- **THEN** coin-api MUST return HTTP 409 Conflict

#### Scenario: Delete not found

- **WHEN** publisher deletes a non-existent component version
- **THEN** coin-api MUST return HTTP 404 Not Found

### Requirement: Delete branching-model draft via Admin API

coin-api SHALL accept delete draft requests for `branching-model` component versions through the generic component delete endpoint.

#### Scenario: Delete branching-model draft succeeds

- **WHEN** publisher calls `DELETE /v1/admin/components/branching-model/trunk-based/versions/2.0.0-draft`
- **AND** the version has `status = draft`
- **THEN** coin-api MUST return HTTP 204 No Content
- **AND** MUST remove the `component_versions` row
- **AND** MUST cascade-delete `component_artifact_bodies` rows for that version (e.g. `model.yaml`)
- **AND** MUST write audit log action `delete_component_draft`
- **AND** MUST NOT modify Nexus

#### Scenario: Reject delete published branching-model

- **WHEN** publisher attempts to delete `branching-model/trunk-based@1.0.0` with `status = published`
- **THEN** coin-api MUST return HTTP 409 Conflict

