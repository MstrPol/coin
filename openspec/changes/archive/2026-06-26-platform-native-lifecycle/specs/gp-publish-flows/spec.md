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

## ADDED Requirements

### Requirement: Composition pin status display

GP draft and release detail views SHALL display component pin lifecycle status alongside type, name, and version.

#### Scenario: Show pin status on release detail

- **WHEN** publisher views GP release detail composition
- **THEN** the UI MUST show status for each pin (`draft` or `published`)
- **AND** MUST link each pin to the corresponding Platform entity page

#### Scenario: Draft pin warning

- **WHEN** a composition pin has `status = draft`
- **THEN** the UI MUST show warning text that the pin may change
- **AND** MUST show the warning on GP draft detail and when assigning GP draft to canary line

### Requirement: GP promote blocked by draft component pins

The UI SHALL prevent GP promote when any composition pin is not `published`.

#### Scenario: Promote button disabled

- **WHEN** publisher views GP draft detail with at least one pin in `draft` status
- **THEN** the UI MUST disable the Promote action
- **AND** MUST list blocking pins with links to publish those component versions on Platform

#### Scenario: Promote API error surfaced

- **WHEN** publisher attempts promote and coin-api returns HTTP 409 for draft pins
- **THEN** the UI MUST display the blocking pin list from the API error payload
