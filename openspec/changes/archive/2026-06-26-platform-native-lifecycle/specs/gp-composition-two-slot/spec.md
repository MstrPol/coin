## ADDED Requirements

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
