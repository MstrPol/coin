## ADDED Requirements

### Requirement: Scoped new release action

New GP releases SHALL be initiated from the GP hub context, not from a global Golden Paths nav entry.

#### Scenario: New release from hub

- **WHEN** publisher clicks «New release» on GP hub for `go-app`
- **THEN** the UI MUST open a release flow scoped to `go-app` without requiring GP selection

### Requirement: Profile create without hidden publish

Creating a new GP profile SHALL NOT automatically publish a release without explicit user confirmation.

#### Scenario: Create profile only

- **WHEN** publisher completes new GP profile creation
- **THEN** the UI MUST create the profile in coin-api and MUST NOT publish a release unless the user explicitly confirms an optional initial release step

### Requirement: Promote draft from release context

Promoting a draft GP release SHALL be available from release list or release detail, not from a global multi-purpose publish wizard in navigation.

#### Scenario: Promote from release detail

- **WHEN** a draft release is open on release detail
- **THEN** the UI MUST offer promote using the existing promote API flow

#### Scenario: No global publish wizard in nav

- **WHEN** publisher views Golden Paths sidebar
- **THEN** the UI MUST NOT show a top-level «Publish» entry pointing to `/releases/publish`

### Requirement: Legacy publish wizard redirect

The global publish wizard route SHALL redirect to GP-scoped release flows when possible.

#### Scenario: Redirect publish wizard with GP query

- **WHEN** user opens `/releases/publish?name=go-app`
- **THEN** the UI MUST redirect to the GP-scoped new release or releases tab for `go-app`
