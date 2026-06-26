## ADDED Requirements

### Requirement: Platform navigation without Jenkins library

The Platform sidebar group SHALL list runtime, build stacks, and branching models only — not Jenkins library management.

#### Scenario: Platform nav items

- **WHEN** user views the Platform group in the sidebar
- **THEN** the UI MUST include Runtime, Build stacks, and Branching models
- **AND** MUST NOT include a Jenkins library entry

### Requirement: Platform settings without runtime lib pin

The Platform settings admin page SHALL configure Nexus integration only and MUST NOT expose lib pin editing.

The page MUST render successfully when the platform settings API response omits a `runtime` field (coin-lib is outside control plane).

#### Scenario: Settings form fields

- **WHEN** admin opens `/platform-settings`
- **THEN** the UI MUST show Nexus configuration fields
- **AND** MUST NOT show platform lib pin fields or runtime section

#### Scenario: Settings without runtime in API response

- **WHEN** admin opens `/platform-settings`
- **AND** `GET /v1/admin/platform/settings` returns only `nexusMavenBase`, `nexusCredentialsId`, and `updatedAt`
- **THEN** the UI MUST render the page without JavaScript errors
- **AND** MUST NOT access or display `runtime.lib`
