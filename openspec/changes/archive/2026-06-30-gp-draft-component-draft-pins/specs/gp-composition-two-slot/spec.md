## ADDED Requirements

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
