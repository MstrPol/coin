## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Canary catalog picker includes GP drafts

The GP Policy tab version picker for `latestCanary` SHALL list GP releases with `status = draft` and `status = published`.

The picker for `latest` (stable) and `minimum` SHALL list only `published` GP releases without snapshot suffix.

#### Scenario: Select GP draft for canary line

- **WHEN** publisher opens GP Policy tab for a profile with GP draft releases
- **THEN** the `Latest canary` dropdown MUST include draft GP versions labeled with `(draft)`
- **AND** saving MUST call catalog update API successfully

#### Scenario: Stable picker excludes drafts

- **WHEN** publisher opens `Latest (stable)` dropdown
- **THEN** the UI MUST NOT list GP versions with `status = draft`

