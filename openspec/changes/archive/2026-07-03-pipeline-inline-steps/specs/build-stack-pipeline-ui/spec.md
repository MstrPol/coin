## ADDED Requirements

### Requirement: Pipeline-first editor layout

coin-ui Build Stack editor SHALL present **Parameters** and **Pipeline stages** as the only primary editing sections. Editor MUST NOT show separate cards for Build targets, Deliverables, or Containerfile artifacts.

#### Scenario: Open v3 draft editor

- **WHEN** publisher opens pipeline-inline draft edit page
- **THEN** UI MUST show Parameters and Pipeline stages sections only
- **AND** MUST NOT show Containerfile artifacts catalog card

#### Scenario: Preview panel placement

- **WHEN** publisher edits pipeline-inline draft
- **THEN** resolved manifest preview MUST appear in right column before Lifecycle
- **AND** preview MUST show containerfile co-located with pipeline steps

### Requirement: Inline step editor by action type

Pipeline editor SHALL provide inline forms per `run`, `build`, `publish` action inside stage cards.

#### Scenario: Buildkit run step includes containerfile editor

- **WHEN** publisher adds or edits `run` step with `engine: buildkit`
- **THEN** UI MUST show textarea for `containerfile.body` inside the same step card
- **AND** MUST NOT require navigating to separate Containerfiles section

#### Scenario: Build step inline fields

- **WHEN** publisher adds `build` step
- **THEN** UI MUST collect `build.id`, `type`, engine fields, containerfile body for buildkit, and publish metadata inline

#### Scenario: Publish step selects build output

- **WHEN** publisher adds `publish` step
- **THEN** UI MUST offer selection of existing `build.id` values from prior build steps

### Requirement: Parameters editor rules

Parameters section SHALL show `allowedValues` only for `type: enum` and `required` checkbox for all types.

#### Scenario: Enum allowed values input

- **WHEN** publisher edits enum allowed values
- **THEN** UI MUST parse comma-separated values on blur without losing comma during typing

### Requirement: Short hash identifiers for stages and build outputs

Pipeline stage `id` and `build.id` SHALL use auto-generated **short hash** identifiers: **5–6 characters**, lowercase `a-z` and `0-9` only (`^[a-z0-9]{5,6}$`). Human-readable labels belong in `stage.name`, not in machine ids.

#### Scenario: UI generates short hash on create

- **WHEN** publisher adds a new pipeline stage or `build` step in Component Studio
- **THEN** UI MUST assign a unique id matching `^[a-z0-9]{5,6}$`
- **AND** publisher MAY edit the id manually while it remains unique and valid

#### Scenario: Validate short hash format

- **WHEN** draft contains `pipeline.stages[].id` or `build.id` outside `^[a-z0-9]{5,6}$`
- **THEN** validate-package MUST reject the draft with a field-level error

#### Scenario: Uniqueness of build.id

- **WHEN** two `build` steps share the same `build.id`
- **THEN** validation MUST reject the draft (unchanged from v3 inline rules)
