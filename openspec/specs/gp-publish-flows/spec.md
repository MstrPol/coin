# gp-publish-flows Specification

## Purpose
TBD - created by archiving change coin-ui-gp-entity-hub. Update Purpose after archive.
## Requirements
### Requirement: Profile create without hidden publish

Creating a new GP profile SHALL create metadata only and SHALL NOT collect component versions or slot bindings.

#### Scenario: Create profile only

- **WHEN** publisher completes new GP profile creation
- **THEN** the UI MUST send only `name` and optional `description` to coin-api and MUST NOT create a draft or release in the same step

### Requirement: Promote draft from release context

Promoting a draft GP release SHALL be available from release list or release detail, not from a global multi-purpose publish wizard in navigation.

#### Scenario: Promote from release detail

- **WHEN** a draft release is open on release detail
- **THEN** the UI MUST offer promote using the existing promote API flow

#### Scenario: No global publish wizard in nav

- **WHEN** publisher views Golden Paths sidebar
- **THEN** the UI MUST NOT show a top-level «Publish» entry pointing to `/releases/publish`

### Requirement: Draft-only operator publish path

The operator UI SHALL NOT offer direct stable GP release publish without an existing draft.

#### Scenario: No direct publish button

- **WHEN** publisher views GP hub or Releases tab
- **THEN** the UI MUST NOT show a «New release» or direct publish action

#### Scenario: Promote is sole publish path

- **WHEN** publisher wants a published GP release
- **THEN** the UI MUST require creating a draft and using promote on draft release detail

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

#### Scenario: Wizard rejects missing published agent

- **WHEN** publisher selects `agent/agent-30-06` with no `published` versions
- **THEN** the agent version dropdown MUST be empty or show only published versions
- **AND** the UI MUST explain that agent pin requires a published version before GP draft can be created

#### Scenario: Wizard warns draft branching pin blocks GP promote

- **WHEN** publisher selects branching-model pin with `status = draft` in the new draft wizard
- **THEN** the UI MUST show a warning that GP promote is blocked until branching-model pin is published
- **AND** MUST NOT reference gp-content publish workflow

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

The UI SHALL prevent GP promote when any external composition pin is not `published` or embedded pipeline is invalid.

#### Scenario: Promote button disabled for draft branching pin

- **WHEN** publisher views GP draft detail with branching-model pin in `draft` status
- **THEN** the UI MUST disable the Promote action
- **AND** MUST list blocking external pins with links to publish those component versions on Platform

#### Scenario: Promote API error surfaced

- **WHEN** publisher attempts promote and coin-api returns HTTP 409 for draft pins
- **THEN** the UI MUST display the blocking pin list from the API error payload

### Requirement: Draft wizard aligned form layout

The draft composition form (create and edit) SHALL use a consistent layout: **Slot**, **Component**, **Version** for the two external pins.

#### Scenario: Column alignment on desktop

- **WHEN** publisher opens new draft composition form on a viewport ≥ `sm`
- **THEN** each row MUST show slot key, component name dropdown, and version dropdown in aligned columns for agent and branching-model only

#### Scenario: Responsive narrow layout

- **WHEN** publisher views the form on a narrow viewport
- **THEN** fields MAY stack vertically while preserving slot → component → version order per row

### Requirement: Draft composition editable until promote

GP draft releases SHALL allow updating composition pins (agent, branching-model) and embedded pipeline until promoted to published.

Published GP releases SHALL NOT allow composition edits.

#### Scenario: Update draft composition

- **WHEN** publisher PATCHes GP draft with new `agentStackName`, `branchingModelName`, and composition versions
- **THEN** coin-api MUST validate and replace `gp_composition` rows for that draft
- **AND** MUST NOT accept `gpContentName`
- **AND** MUST write audit log action `update_gp_draft`
- **AND** MUST NOT change the draft version string

#### Scenario: Update published rejected

- **WHEN** publisher attempts to PATCH GP release `xxx@1.0.0` with `status = published`
- **THEN** coin-api MUST reject with HTTP 409 Conflict

#### Scenario: Edit draft from release detail

- **WHEN** publisher opens release detail for a draft
- **THEN** the UI MUST show editable composition controls for two external pins
- **AND** MUST show embedded pipeline editor
- **AND** MUST offer **Save** to persist composition and pipeline changes
- **AND** MUST offer **Delete draft** and **Promote**

#### Scenario: Published release detail read-only

- **WHEN** publisher opens release detail for a published release
- **THEN** the UI MUST show composition as read-only
- **AND** MUST NOT offer Save or Delete actions

### Requirement: Draft deletion

GP draft releases (`status = draft`) SHALL be deletable by publishers. Published GP releases (`status = published`) SHALL NOT be deletable through operator API or UI because they are immutable in Nexus after promote.

#### Scenario: Delete draft succeeds

- **WHEN** publisher deletes GP draft `go-app@1.0.0-snapshot.2` with `status = draft`
- **THEN** coin-api MUST remove the draft from PostgreSQL (release row, composition, draft artifacts)
- **AND** MUST write audit log action `delete_gp_draft`
- **AND** MUST NOT modify Nexus published blobs

#### Scenario: Delete published rejected

- **WHEN** publisher attempts to delete GP release `go-app@1.0.0` with `status = published`
- **THEN** coin-api MUST reject with HTTP 409 Conflict
- **AND** the UI MUST NOT offer a delete action for published releases

#### Scenario: Delete draft from hub

- **WHEN** publisher views a draft on Releases tab or release detail
- **THEN** the UI MUST show «Delete draft» (publisher+)
- **AND** MUST confirm before delete

### Requirement: Legacy publish wizard redirect

The legacy publish wizard route SHALL redirect to GP-scoped draft creation.

#### Scenario: Redirect publish wizard with GP query

- **WHEN** user opens `/releases/publish?name=go-app`
- **THEN** the UI MUST redirect to `/gp/go-app/releases/new-draft`

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

