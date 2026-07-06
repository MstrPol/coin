## MODIFIED Requirements

### Requirement: Release detail shows version composition

GP release detail SHALL show external composition pins for that version and embedded pipeline authoring controls on draft releases. Composition table MUST list agent and branching-model pins only and MUST NOT reference gp-content or Platform build stacks.

#### Scenario: Published release composition table

- **WHEN** enabling team opens release detail for GP `go-app` version `1.0.0`
- **THEN** the UI MUST show the composition table for **that version** with agent and branching-model pins only
- **AND** agent pin MUST link to `/platform/runtime/{agentName}/releases/{version}`
- **AND** branching-model pin MUST link to `/platform/branching-models/{name}/releases/{version}` or edit route for draft
- **AND** MUST NOT show gp-content composition row or link to `/platform/build-stacks`

#### Scenario: Draft release detail includes pipeline editor

- **WHEN** publisher opens release detail for a GP draft
- **THEN** the UI MUST offer embedded pipeline editor (Parameters + Pipeline stages) on the same page
- **AND** MUST offer promote and delete draft actions
- **AND** MUST NOT offer delete for published releases

## ADDED Requirements

### Requirement: GP release detail pipeline tab

GP release detail for draft releases SHALL be the primary authoring surface for embedded pipeline-inline model.

#### Scenario: Pipeline section on draft release detail

- **WHEN** publisher opens GP draft release detail
- **THEN** the UI MUST show pipeline editor section before or alongside composition pins
- **AND** MUST call GP release pipeline preview API on edit

#### Scenario: Published release pipeline read-only

- **WHEN** publisher opens published GP release detail
- **THEN** the UI MUST show pipeline content as read-only
- **AND** MUST NOT offer pipeline save controls
