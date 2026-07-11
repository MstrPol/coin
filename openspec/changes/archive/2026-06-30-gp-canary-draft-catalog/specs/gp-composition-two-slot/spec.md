## MODIFIED Requirements

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
