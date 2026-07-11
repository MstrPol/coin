## MODIFIED Requirements

### Requirement: Manual catch-up draft form

When creating an agent draft from Platform UI (image already in registry), the form SHALL require image ref and digest only.

The form MUST derive and display the version as a read-only preview parsed from the image tag.

The form MUST NOT include an editable Version field or GOARCH / architecture fields.

#### Scenario: Create catch-up draft

- **WHEN** publisher opens «New draft» for an agent profile without CI register
- **THEN** the UI MUST require Image ref and Digest before submit
- **AND** MUST show read-only derived version from the image tag (e.g. `1.2.0`)
- **AND** MUST NOT show editable Version or GOARCH inputs

#### Scenario: Disable submit when tag unparseable

- **WHEN** publisher enters an image ref without a valid tag
- **THEN** the UI MUST disable submit
- **AND** MUST show validation hint on Image ref

#### Scenario: CI-registered draft release detail

- **WHEN** publisher opens release detail for agent draft created by CI
- **THEN** the UI MUST display image and digest read-only
- **AND** MUST offer Promote action for publisher role
- **AND** MUST NOT offer GOARCH field
