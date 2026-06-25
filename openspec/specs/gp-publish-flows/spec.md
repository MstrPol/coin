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

- **WHEN** publisher clicks «New draft» on GP hub for `xxx`
- **THEN** the UI MUST open a draft composition flow scoped to GP `xxx`
- **AND** MUST offer catalog pickers for **agent stack**, **branching-model**, and **gp-content** (name + published version each)

#### Scenario: Draft wizard three components

- **WHEN** publisher opens new draft
- **THEN** the UI MUST show three composition pickers: Agent/executor stack, Branching model, GP content
- **AND** MUST NOT require gp-content name to match GP profile name

### Requirement: Draft wizard aligned form layout

The draft composition form (create and edit) SHALL use a consistent three-column layout: **Slot**, **Component**, **Version**.

#### Scenario: Column alignment on desktop

- **WHEN** publisher opens new draft composition form on a viewport ≥ `sm`
- **THEN** each row MUST show slot key, component name dropdown, and version dropdown in aligned columns
- **AND** each dropdown MUST have a visible label above the control (not inline text pushing version column out of alignment)

#### Scenario: Responsive narrow layout

- **WHEN** publisher views the form on a narrow viewport
- **THEN** fields MAY stack vertically while preserving slot → component → version order per row

### Requirement: Draft composition editable until promote

GP draft releases SHALL allow updating composition pins (agent, gp-content, branching-model) until promoted to published.

Published GP releases SHALL NOT allow composition edits.

#### Scenario: Update draft composition

- **WHEN** publisher PATCHes GP draft `xxx@1.0.0-snapshot.2` with new `agentStackName`, `gpContentName`, `branchingModelName`, and composition versions
- **THEN** coin-api MUST validate and replace `gp_composition` rows for that draft
- **AND** MUST write audit log action `update_gp_draft`
- **AND** MUST NOT change the draft version string

#### Scenario: Update published rejected

- **WHEN** publisher attempts to PATCH GP release `xxx@1.0.0` with `status = published`
- **THEN** coin-api MUST reject with HTTP 409 Conflict

#### Scenario: Edit draft from release detail

- **WHEN** publisher opens release detail for a draft
- **THEN** the UI MUST show editable composition controls (same 3-pin pickers as new draft)
- **AND** MUST offer **Save** to persist changes
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
