# gp-composition-two-slot Specification

## Purpose

GP release composition: two operator pins (`agent`, `branching-model`) plus embedded pipeline on GP release body. Capability id `gp-composition-two-slot` retained for delta traceability. See also `gp-release-two-pin` and `gp-embedded-pipeline`. Narrative docs cross-link `docs/adr/coin-ci-runtime.md`.

## Requirements

### Requirement: Agent pin published only in GP composition

GP draft and published composition SHALL require the `agent` pin to reference a `published` component version.

`branching-model` pin MAY reference `draft` or `published` versions in GP drafts and canary resolve contexts.

#### Scenario: Reject draft agent pin

- **WHEN** publisher creates or updates GP draft with agent pin referencing a non-published version
- **THEN** coin-api MUST reject with a validation error

### Requirement: Promote validates published pins

GP promote SHALL re-validate all external composition pins as `published` and embedded pipeline as valid before transitioning the GP release.

#### Scenario: Promote rejects draft branching-model pin

- **WHEN** publisher promotes GP draft pinning `branching-model/trunk-based@2.0.0-draft` with `status = draft`
- **THEN** coin-api MUST reject promote with HTTP 409 Conflict

#### Scenario: Promote accepts all published pins

- **WHEN** all external composition pins are `published` and embedded pipeline is valid at promote time
- **THEN** coin-api MUST complete GP draft promotion

### Requirement: Canary line GP draft resolve

coin-api SHALL allow canary channel resolve when `catalog_policy.latest_canary` references a GP release with `status = draft`.

coin-api SHALL accept catalog policy updates where `latestCanary` references any existing GP release version with `status` in `draft` or `published`, including snapshot semver suffixes (`-snapshot.N`).

#### Scenario: Canary resolves GP draft

- **WHEN** pilot project resolves GP with canary channel
- **AND** `latest_canary` points to GP draft `2.0.0-snapshot.3`
- **THEN** coin-api MUST resolve that GP draft
- **AND** MUST allow `draft` status for `branching-model` pin per canary resolve rules

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

### Requirement: Documentation consistency for two-pin composition

GP composition documentation in `docs/architecture.md`, `docs/control-plane.md`, and `docs/golden-paths.md` SHALL remain consistent with the two-pin composition requirements in this specification and SHALL cross-link `docs/adr/coin-ci-runtime.md` and `docs/adr/gp-embedded-pipeline.md`.

#### Scenario: Architecture references composition spec

- **WHEN** `docs/architecture.md` describes GP composition slots
- **THEN** it MUST list exactly `agent` and `branching-model` as external pins
- **AND** MUST describe embedded pipeline as GP release payload (not a composition pin)
- **AND** MUST link to `docs/adr/coin-ci-runtime.md` for runtime slot materialization

#### Scenario: Golden paths doc matches two-pin

- **WHEN** reader opens `docs/golden-paths.md` for composition rules
- **THEN** the document MUST describe pins `agent` and `branching-model` only
- **AND** MUST state that pipeline-inline lives on the GP release body
- **AND** MUST NOT present `gp-content` as a required composition pin for new releases
