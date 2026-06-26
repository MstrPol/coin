# gp-composition-two-slot Specification

## Purpose
TBD - created by archiving change gp-profile-metadata-model. Capability id retained for delta traceability; composition is three-pin (agent, branching-model, gp-content).

## Requirements

### Requirement: Three-pin GP draft composition

GP draft and release composition SHALL contain exactly three operator-selected component pins from the platform registry:

1. **`agent`** — CI runtime stack (agent / executor; e.g. `coin-agent`)
2. **`branching-model`** — versioning and publish policy
3. **`gp-content`** — build stack (Containerfile, schema, pipeline)

Standalone `executor` SHALL NOT appear in the GP composition map; resolve materializes executor from the selected agent stack.

coin-api SHALL NOT inject Jenkins Shared Library (`lib`) from platform settings or any other control-plane source during resolve.

#### Scenario: Create draft with three catalog pins

- **WHEN** publisher creates a draft with `agentStackName` `coin-agent`, `gpContentName` `go-app`, `branchingModelName` `trunk-based`, and matching versions in the composition map
- **THEN** coin-api MUST accept the draft
- **AND** MUST require keys `agent`, `gp-content`, and `branching-model` in composition

#### Scenario: Reject lib in GP draft composition

- **WHEN** publisher attempts to create a draft including `lib` in composition
- **THEN** coin-api MUST reject the request with a validation error

#### Scenario: Reject standalone executor in GP draft composition

- **WHEN** publisher attempts to create a draft including `executor` as a separate composition key
- **THEN** coin-api MUST reject the request with a validation error

#### Scenario: Resolve without lib injection

- **WHEN** resolve runs for a published GP release with three-pin composition
- **THEN** coin-api MUST materialize executor from the pinned agent stack
- **AND** MUST NOT add `lib` to the resolved manifest

### Requirement: gp-content pinned per GP version not profile

`gp-content` (build stack) SHALL be pinned only in GP **version** composition (`gp_composition` per draft/release), not on the GP profile entity.

#### Scenario: No profile-level gp-content binding

- **WHEN** enabling team views GP profile `xxx` without opening a specific release
- **THEN** coin-api and coin-ui MUST NOT expose a profile → build stack relationship or implied default gp-content for the profile

#### Scenario: Different releases may pin different gp-content

- **WHEN** GP `xxx` has release `1.0.0` with `gp-content/go-app@1.0.0` and draft `2.0.0-snapshot.1` with `gp-content/other-stack@1.0.0`
- **THEN** each version MUST retain its own composition pins independently

### Requirement: Component catalog independence

`agent`, `gp-content`, and `branching-model` components SHALL exist in the platform registry independently of any GP profile.

#### Scenario: gp-content without matching profile

- **WHEN** `gp-content/go-app@1.0.0` is published in the component registry
- **AND** no GP profile named `go-app` exists
- **THEN** the component MUST remain valid and selectable for future GP drafts

#### Scenario: GP profile without matching gp-content

- **WHEN** GP profile `xxx` exists
- **AND** no `gp-content/xxx` component exists
- **THEN** publisher MUST still be able to create a draft by selecting another gp-content component (e.g. `go-app`) from the catalog

### Requirement: Explicit component names in draft API

New draft and publish requests SHALL include explicit component names separate from the GP profile name.

#### Scenario: agentStackName gpContentName branchingModelName required

- **WHEN** publisher creates a draft
- **THEN** the request MUST include `agentStackName`, `gpContentName`, and `branchingModelName`
- **AND** coin-api MUST validate versions against those component names

#### Scenario: Profile name may differ from gp-content name

- **WHEN** publisher creates a draft for GP profile `xxx` with `gpContentName` `go-app`
- **THEN** coin-api MUST store composition referencing `gp-content/go-app` at the pinned version
- **AND** resolve for GP `xxx` MUST materialize manifest using that gp-content bundle and the selected agent stack

### Requirement: Agent pin published only in GP composition

GP draft and published composition SHALL require the `agent` pin to reference a `published` component version.

`gp-content` and `branching-model` pins MAY reference `draft` or `published` versions in GP drafts and canary resolve contexts.

#### Scenario: Reject draft agent pin

- **WHEN** publisher creates or updates GP draft with agent pin referencing a non-published version
- **THEN** coin-api MUST reject with a validation error

#### Scenario: Accept draft gp-content in GP draft

- **WHEN** publisher creates GP draft with `gp-content/go-app@1.2.0-draft` where that version has `status = draft`
- **THEN** coin-api MUST accept the draft
- **AND** MUST store the composition pin

### Requirement: Promote validates published pins

GP promote SHALL re-validate all composition pins as `published` before transitioning the GP release.

#### Scenario: Promote rejects draft branching-model pin

- **WHEN** publisher promotes GP draft pinning `branching-model/trunk-based@2.0.0-draft` with `status = draft`
- **THEN** coin-api MUST reject promote with HTTP 409 Conflict

#### Scenario: Promote accepts all published pins

- **WHEN** all composition pins are `published` at promote time
- **THEN** coin-api MUST complete GP draft promotion

### Requirement: Canary line GP draft resolve

coin-api SHALL allow canary channel resolve when `catalog_policy.latest_canary` references a GP release with `status = draft`.

#### Scenario: Canary resolves GP draft

- **WHEN** pilot project resolves GP with canary channel
- **AND** `latest_canary` points to GP draft `2.0.0-snapshot.3`
- **THEN** coin-api MUST resolve that GP draft
- **AND** MUST allow `draft` status for `gp-content` and `branching-model` pins per canary resolve rules

#### Scenario: Stable rejects GP draft

- **WHEN** product CI resolves GP on stable channel
- **THEN** coin-api MUST NOT resolve GP releases with `status = draft`
