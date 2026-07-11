## MODIFIED Requirements

### Requirement: Draft creation from hub

New GP composition work SHALL be initiated only through draft creation from the GP hub context.

#### Scenario: New draft from hub

- **WHEN** publisher clicks «New draft» on GP hub for `xxx`
- **THEN** the UI MUST open a draft composition flow scoped to GP `xxx`
- **AND** MUST offer catalog pickers for **agent stack** (published versions only), **branching-model**, and **gp-content** (draft and published versions)

#### Scenario: Draft wizard three components

- **WHEN** publisher opens new draft
- **THEN** the UI MUST show three composition pickers: Agent/executor stack, Branching model, GP content
- **AND** MUST display status badge for each selected version (`draft` or `published`)
- **AND** MUST NOT require gp-content name to match GP profile name

#### Scenario: Wizard lists draft component versions

- **WHEN** publisher selects `gp-content/bs-30-06` that has only `status = draft` versions
- **THEN** the version dropdown MUST list those draft versions with a `(draft)` label
- **AND** MUST allow creating the GP draft with the selected draft pin

#### Scenario: Wizard rejects missing published agent

- **WHEN** publisher selects `agent/agent-30-06` with no `published` versions
- **THEN** the agent version dropdown MUST be empty or show only published versions
- **AND** the UI MUST explain that agent pin requires a published version before GP draft can be created

#### Scenario: Wizard warns draft pins block GP promote

- **WHEN** publisher selects at least one composition pin with `status = draft` in the new draft wizard
- **THEN** the UI MUST show a warning that GP promote is blocked until all pins are published
- **AND** MUST NOT imply that component drafts will auto-publish on GP promote

## ADDED Requirements

### Requirement: GP draft wizard version picker parity

The new GP draft wizard SHALL use the same component version visibility rules as the GP draft composition editor on release detail.

#### Scenario: gp-content and branching-model include drafts

- **WHEN** publisher changes gp-content or branching-model component name in the new draft wizard
- **THEN** the UI MUST load versions with `status = draft` and `status = published`
- **AND** MUST NOT filter those slots to published-only

#### Scenario: Agent remains published-only in wizard

- **WHEN** publisher changes agent stack name in the new draft wizard
- **THEN** the UI MUST load only `status = published` versions for the agent slot
