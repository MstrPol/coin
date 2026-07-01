# gp-composition-two-slot Specification

## Purpose

GP release composition: three operator pins (`agent`, `gp-content`, `branching-model`). Capability id `gp-composition-two-slot` retained for delta traceability. Narrative docs cross-link `docs/adr/coin-ci-runtime.md`.
## Requirements
### Requirement: Three-pin GP draft composition

GP release composition SHALL contain exactly three operator-controlled pins:

1. **`agent`** — CI runtime stack (container image with baked `coin-executor`; e.g. `coin-agent`, `agent-30-06`)
2. **`gp-content`** — build policy, Containerfile, schema
3. **`branching-model`** — versioning and publish policy

Standalone `executor` SHALL NOT appear in the GP composition map, registry, or resolved manifest. The agent pin is the sole runtime source of truth.

Resolved manifest SHALL be a deterministic materialization of the GP release identity and these three pins only. `coin-api` SHALL NOT add site-local Jenkins glue fields, credential IDs, or synthetic runtime sections that are not sourced from the GP release identity, `agent`, `gp-content`, or `branching-model`.

#### Scenario: Create GP draft with three pins

- **WHEN** publisher creates GP draft with agent, gp-content, and branching-model versions
- **THEN** coin-api MUST persist exactly three composition rows
- **AND** MUST NOT validate or require `executor/coin-executor@{agentVersion}` in the component registry

#### Scenario: Reject standalone executor in GP draft composition

- **WHEN** publisher attempts to create a draft including `executor` as a separate composition key
- **THEN** coin-api MUST reject with invalid composition error

#### Scenario: Resolve materializes runtime from agent only

- **WHEN** GP release is resolved for CI
- **THEN** coin-api MUST populate `manifest.runtime` from the pinned agent version metadata
- **AND** MUST NOT add `manifest.executor`
- **AND** MUST NOT query component registry for type `executor`

#### Scenario: Resolve emits only composition-owned sections

- **WHEN** GP release `gp-01-07@1.0.0` is resolved for CI
- **THEN** the manifest MUST contain GP identity fields `goldenPath.name` and `goldenPath.version`
- **AND** MUST contain `runtime` materialized from the `agent` pin
- **AND** MUST contain `build`, `pipeline`, `validateSchema`, and `capabilities` materialized from the `gp-content` pin
- **AND** MUST contain `branching` materialized from the `branching-model` pin
- **AND** MUST preserve resolve integrity metadata `manifestVersion` and `manifestHash`
- **AND** MUST NOT contain top-level `credentials`, `lib`, `executor`, or any Jenkins-instance credential ID

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

coin-api SHALL accept catalog policy updates where `latestCanary` references any existing GP release version with `status` in `draft` or `published`, including snapshot semver suffixes (`-snapshot.N`).

#### Scenario: Canary resolves GP draft

- **WHEN** pilot project resolves GP with canary channel
- **AND** `latest_canary` points to GP draft `2.0.0-snapshot.3`
- **THEN** coin-api MUST resolve that GP draft
- **AND** MUST allow `draft` status for `gp-content` and `branching-model` pins per canary resolve rules

#### Scenario: Stable rejects GP draft

- **WHEN** product CI resolves GP on stable channel
- **THEN** coin-api MUST NOT resolve GP releases with `status = draft`

#### Scenario: Assign GP draft to canary catalog line

- **WHEN** publisher PATCHes catalog policy with `latestCanary` set to an existing GP draft version
- **THEN** coin-api MUST accept the update
- **AND** MUST NOT require that version to be `published`

#### Scenario: Stable catalog fields remain published-only

- **WHEN** publisher PATCHes catalog policy `latest` or `minimum`
- **THEN** coin-api MUST require referenced GP versions to be `published`
- **AND** MUST reject snapshot semver for those fields

### Requirement: Composition documentation cross-links

GP composition documentation in `docs/architecture.md` and `docs/control-plane.md` SHALL remain consistent with the three-pin composition requirements in this specification and SHALL cross-link `docs/adr/coin-ci-runtime.md`.

#### Scenario: Architecture references composition spec

- **WHEN** `docs/architecture.md` describes GP composition slots
- **THEN** it MUST list exactly `agent`, `gp-content`, and `branching-model`
- **AND** MUST link to `docs/adr/coin-ci-runtime.md` for runtime slot materialization

### Requirement: GP draft on draft component pins with promote gate

The platform SHALL allow creating and editing GP drafts that pin draft `gp-content` and/or `branching-model` versions while preventing GP promote until every composition pin is `published`.

#### Scenario: Create GP draft with draft build stack and published agent

- **WHEN** publisher creates GP draft with `agent/coin-agent@1.0.0` (`published`), `gp-content/bs-30-06@0.1.0-draft` (`draft`), and `branching-model/bm-30-06@0.1.0-draft` (`draft`)
- **THEN** coin-api MUST accept the GP draft
- **AND** the UI MUST allow saving that composition

#### Scenario: Promote GP draft blocked by any draft component pin

- **WHEN** publisher attempts to promote a GP draft where any composition pin has `status = draft` (including `gp-content` or `branching-model`)
- **THEN** coin-api MUST reject promote with HTTP 409 Conflict
- **AND** the UI MUST disable or block promote and list blocking pins

#### Scenario: Promote GP draft after all pins published

- **WHEN** publisher promotes GP draft after every composition pin has `status = published`
- **THEN** coin-api MUST complete GP promote to `published`

