## MODIFIED Requirements

### Requirement: GP content schema v2 editor

The build stack editor SHALL edit Build Stack vNext canonical model bijectively with ordered section cards: parameters, build targets, deliverables, Containerfile artifacts, pipeline stages and manifest preview. Raw `content.yaml` SHALL NOT be the primary editing surface.

#### Scenario: Target card switches engine block

- **WHEN** publisher selects engine `dockerfile` for a target in the editor
- **THEN** the UI MUST show BYO Dockerfile fields for that target
- **AND** MUST NOT hide other targets that use managed Containerfile artifacts

#### Scenario: Target card buildkit

- **WHEN** publisher selects engine `buildkit` for a target
- **THEN** the UI MUST show buildkit target fields and managed Containerfile artifact selector for that target
- **AND** MUST allow that target to be referenced by deliverables

#### Scenario: Save produces vNext model

- **WHEN** publisher saves draft from the editor
- **THEN** persisted gp-content package MUST contain Build Stack vNext model
- **AND** MUST NOT persist conflicting raw YAML as a second source of truth

### Requirement: Build stack preview panel

The build stack editor SHALL call gp-content preview API and display resolved manifest preview, validation issues and warnings from Build Stack vNext canonical model.

#### Scenario: Debounced preview on edit

- **WHEN** publisher changes parameters, targets, deliverables, Containerfiles or stages in draft editor
- **THEN** the UI MUST call preview API and show resolved manifest preview
- **AND** MUST NOT show duplicated YAML and JSON previews as equal primary views
