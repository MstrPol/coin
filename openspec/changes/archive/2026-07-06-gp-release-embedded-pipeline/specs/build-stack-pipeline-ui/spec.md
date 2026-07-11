## MODIFIED Requirements

### Requirement: Pipeline-first editor layout

coin-ui GP release pipeline editor SHALL present **Parameters** and **Pipeline stages** as the only primary editing sections on GP release detail for draft releases. Editor MUST NOT show separate cards for Build targets, Deliverables, or Containerfile artifacts. Editor MUST NOT live under Platform → Build stacks routes.

#### Scenario: Open pipeline on GP draft release detail

- **WHEN** publisher opens GP draft release detail for `go-app@1.0.0-snapshot.1`
- **THEN** UI MUST show Parameters and Pipeline stages sections on the release detail page
- **AND** MUST NOT require navigation to `/platform/build-stacks`

#### Scenario: Preview panel on release detail

- **WHEN** publisher edits embedded pipeline on GP draft release detail
- **THEN** resolved manifest preview MUST appear alongside the editor
- **AND** preview MUST show containerfile co-located with pipeline steps

### Requirement: Inline step editor by action type

Pipeline editor on GP release detail SHALL provide inline forms per `run`, `build`, `publish` action inside stage cards.

#### Scenario: Buildkit run step includes containerfile editor

- **WHEN** publisher adds or edits `run` step with `engine: buildkit` on GP release pipeline editor
- **THEN** UI MUST show textarea for `containerfile.body` inside the same step card
- **AND** MUST NOT require navigating to a separate Platform build stacks page

#### Scenario: Publish step selects build output

- **WHEN** publisher adds `publish` step on GP release pipeline editor
- **THEN** UI MUST offer selection of existing `build.id` values from prior build steps in the same release pipeline

### Requirement: Short hash identifiers for stages and build outputs

Pipeline stage `id` and `build.id` on GP release pipeline editor SHALL use auto-generated **short hash** identifiers: **5–6 characters**, lowercase `a-z` and `0-9` only (`^[a-z0-9]{5,6}$`).

#### Scenario: UI generates short hash on create

- **WHEN** publisher adds a new pipeline stage or `build` step on GP release detail
- **THEN** UI MUST assign a unique id matching `^[a-z0-9]{5,6}$`

#### Scenario: Validate short hash format on save

- **WHEN** GP release pipeline draft contains `pipeline.stages[].id` or `build.id` outside `^[a-z0-9]{5,6}$`
- **THEN** validation MUST reject the draft with a field-level error
