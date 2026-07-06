## ADDED Requirements

### Requirement: Build Stack vNext editor sections

coin-ui SHALL provide Build Stack vNext editor with separate sections for parameters, build targets, deliverables, Containerfile artifacts, pipeline stages and preview.

#### Scenario: Open editor

- **WHEN** publisher opens Build Stack draft edit page
- **THEN** UI MUST show structured cards for parameters, targets, deliverables, Containerfiles, pipeline stages and manifest preview
- **AND** UI MUST NOT make raw YAML the primary editing surface

### Requirement: Dynamic deliverables editor

coin-ui SHALL allow publisher to dynamically add, remove and configure deliverables.

#### Scenario: Add image deliverable

- **WHEN** publisher adds deliverable of type `image`
- **THEN** UI MUST require deliverable id, target reference and publish settings
- **AND** UI MUST show validation if referenced target is missing

### Requirement: Dynamic target editor

coin-ui SHALL allow publisher to dynamically add, remove and configure build targets with target-level engine selection.

#### Scenario: Configure target engine

- **WHEN** publisher selects engine `buildkit` for a target
- **THEN** UI MUST show buildkit-specific fields
- **AND** MUST allow selecting a managed Containerfile artifact

#### Scenario: Configure dockerfile target

- **WHEN** publisher selects engine `dockerfile` for a target
- **THEN** UI MUST show workspace Dockerfile path fields
- **AND** MUST NOT require managed Containerfile artifact

### Requirement: Parameters editor

coin-ui SHALL allow publisher to create typed non-secret parameters and reference them from supported fields.

#### Scenario: Add enum parameter

- **WHEN** publisher creates enum parameter
- **THEN** UI MUST require allowed values
- **AND** preview MUST show the parameter in manifest preview

### Requirement: Non-duplicating preview

coin-ui SHALL show one canonical resolved manifest preview and validation issues. Raw YAML export/debug SHALL NOT be available in P0 Build Stack vNext editor.

#### Scenario: Preview after edit

- **WHEN** publisher changes deliverable or target configuration
- **THEN** UI MUST refresh manifest preview
- **AND** UI MUST NOT show duplicated YAML/JSON representations of the same source model as primary content
