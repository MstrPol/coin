## ADDED Requirements

### Requirement: Build Stack vNext canonical model

Build Stack vNext SHALL use a structured canonical model containing parameters, build targets, deliverables, Containerfile artifacts and pipeline stages. Raw YAML SHALL NOT be the primary editing contract in coin-ui.

#### Scenario: Save canonical model

- **WHEN** publisher saves a Build Stack vNext draft from the editor
- **THEN** coin-api MUST persist a structured gp-content package model
- **AND** the package MUST be sufficient to materialize the resolved manifest without reading product repository build configuration

#### Scenario: Reject duplicate source of truth

- **WHEN** a draft update contains both structured model fields and a conflicting raw YAML body
- **THEN** coin-api MUST reject the update with validation error

### Requirement: Build Stack parameters

Build Stack vNext SHALL define typed non-secret parameters that can be referenced by build targets and pipeline stages.

#### Scenario: Define parameter

- **WHEN** publisher adds parameter `GO_VERSION` with type `string` and default value
- **THEN** preview MUST include that parameter in manifest preview
- **AND** validation MUST reject an empty required parameter without default

#### Scenario: Reject credential parameter

- **WHEN** publisher adds a parameter intended to carry credential ID or secret value
- **THEN** validation MUST reject the parameter
- **AND** the error MUST explain that credentials stay in Jenkins glue/product config

### Requirement: Build Stack targets

Build Stack vNext SHALL define named build targets. Each target SHALL choose its own engine and engine-specific configuration.

#### Scenario: Multiple target engines

- **WHEN** Build Stack defines `app-image` with engine `buildkit` and `migrations-image` with engine `dockerfile`
- **THEN** preview MUST materialize both targets
- **AND** executor MUST be able to dispatch each target by its own engine

### Requirement: Build Stack deliverables

Build Stack vNext SHALL define named deliverables with type, target reference and publish metadata. Deliverables SHALL NOT be represented only as a flat capabilities list.

#### Scenario: Deliverable references target

- **WHEN** publisher creates deliverable `app` of type `image` with `targetId: app-image`
- **THEN** validation MUST require target `app-image` to exist
- **AND** manifest preview MUST include deliverable `app`

#### Scenario: Reject orphan deliverable

- **WHEN** publisher creates a deliverable that references missing target
- **THEN** validation MUST reject the Build Stack draft

### Requirement: Build Stack Containerfile artifacts

Build Stack vNext SHALL support named managed Containerfile artifacts stored in the gp-content package and referenced from targets.

#### Scenario: Target uses managed Containerfile

- **WHEN** target `app-image` references Containerfile artifact `app`
- **THEN** preview MUST include immutable content ref for artifact `app`
- **AND** executor MUST materialize that Containerfile before building the target

#### Scenario: Reject missing Containerfile artifact

- **WHEN** target references a Containerfile artifact id that is not in the package
- **THEN** validation MUST reject the draft
