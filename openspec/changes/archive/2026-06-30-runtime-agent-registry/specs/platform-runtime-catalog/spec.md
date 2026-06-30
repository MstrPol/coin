## MODIFIED Requirements

### Requirement: Runtime publish guidance

The runtime catalog and hub SHALL support draft registration from CI and manual catch-up, with promote as a publisher-only Platform action.

#### Scenario: Show publish runbook hint

- **WHEN** enabling team views runtime catalog or agent hub
- **THEN** the UI MUST show a link or hint to the agent publish runbook for CI image push and API draft register
- **AND** MUST state that promote happens in Platform UI after review
- **AND** MUST NOT reference Component Studio

#### Scenario: New agent profile from catalog

- **WHEN** publisher clicks «New profile» on `/platform/runtime`
- **THEN** the UI MUST navigate to `/platform/runtime/new` to create an agent component profile
- **AND** after create MUST land on hub with welcome CTA for first draft

#### Scenario: Draft agent versions visible in hub

- **WHEN** enabling team opens agent stack hub Releases tab
- **THEN** the UI MUST list both `draft` and `published` versions for that profile
- **AND** MUST offer «New draft» on the hub for additional versions

## ADDED Requirements

### Requirement: Manual catch-up draft form

When creating an agent draft from Platform UI (image already in registry), the form SHALL require version, image ref, and digest.

The form MUST NOT include GOARCH or architecture fields.

#### Scenario: Create catch-up draft

- **WHEN** publisher opens «New draft» for an agent profile without CI register
- **THEN** the UI MUST require Version, Image ref, and Digest before submit
- **AND** MUST NOT show GOARCH input

#### Scenario: CI-registered draft release detail

- **WHEN** publisher opens release detail for agent draft created by CI
- **THEN** the UI MUST display image and digest read-only
- **AND** MUST offer Promote action for publisher role
- **AND** MUST NOT offer GOARCH field
