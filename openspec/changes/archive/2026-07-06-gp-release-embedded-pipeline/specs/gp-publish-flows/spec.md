## MODIFIED Requirements

### Requirement: Draft creation from hub

New GP composition work SHALL be initiated only through draft creation from the GP hub context.

#### Scenario: New draft from hub

- **WHEN** publisher clicks «New draft» on GP hub for `go-app`
- **THEN** the UI MUST open a draft composition flow scoped to GP `go-app`
- **AND** MUST offer catalog pickers for **agent stack** (published versions only) and **branching-model** (draft and published versions)
- **AND** MUST scaffold embedded pipeline for the GP profile on draft creation

#### Scenario: Draft wizard two external components

- **WHEN** publisher opens new draft
- **THEN** the UI MUST show two composition pickers: Agent/executor stack and Branching model
- **AND** MUST NOT show gp-content picker
- **AND** MUST display status badge for each selected version (`draft` or `published`)

#### Scenario: Wizard warns draft branching pin blocks GP promote

- **WHEN** publisher selects branching-model pin with `status = draft` in the new draft wizard
- **THEN** the UI MUST show a warning that GP promote is blocked until branching-model pin is published
- **AND** MUST NOT reference gp-content publish workflow

### Requirement: Draft wizard aligned form layout

The draft composition form (create and edit) SHALL use a consistent layout: **Slot**, **Component**, **Version** for the two external pins.

#### Scenario: Column alignment on desktop

- **WHEN** publisher opens new draft composition form on a viewport ≥ `sm`
- **THEN** each row MUST show slot key, component name dropdown, and version dropdown in aligned columns for agent and branching-model only

### Requirement: Draft composition editable until promote

GP draft releases SHALL allow updating composition pins (agent, branching-model) and embedded pipeline until promoted to published.

#### Scenario: Update draft composition

- **WHEN** publisher PATCHes GP draft with new `agentStackName`, `branchingModelName`, and composition versions
- **THEN** coin-api MUST validate and replace `gp_composition` rows for that draft
- **AND** MUST NOT accept `gpContentName`

#### Scenario: Edit draft from release detail

- **WHEN** publisher opens release detail for a draft
- **THEN** the UI MUST show editable composition controls for two external pins
- **AND** MUST show embedded pipeline editor
- **AND** MUST offer **Save** to persist composition and pipeline changes

### Requirement: GP promote blocked by draft component pins

The UI SHALL prevent GP promote when any external composition pin is not `published` or embedded pipeline is invalid.

#### Scenario: Promote button disabled for draft branching pin

- **WHEN** publisher views GP draft detail with branching-model pin in `draft` status
- **THEN** the UI MUST disable the Promote action
- **AND** MUST list blocking external pins with links to publish those component versions on Platform

### Requirement: Draft creation without platform lib prerequisite

Creating a GP draft SHALL NOT require a configured platform lib pin.

#### Scenario: New draft without platform runtime

- **WHEN** publisher creates a new GP draft with valid agent and branching-model pins
- **THEN** coin-api MUST accept the draft and scaffold embedded pipeline body
- **AND** MUST NOT require gp-content pins

### Requirement: GP draft wizard version picker parity

The new GP draft wizard SHALL use the same component version visibility rules as the GP draft composition editor on release detail.

#### Scenario: Branching-model includes drafts

- **WHEN** publisher changes branching-model component name in the new draft wizard
- **THEN** the UI MUST load versions with `status = draft` and `status = published`

#### Scenario: Agent remains published-only in wizard

- **WHEN** publisher changes agent stack name in the new draft wizard
- **THEN** the UI MUST load only `status = published` versions for the agent slot
