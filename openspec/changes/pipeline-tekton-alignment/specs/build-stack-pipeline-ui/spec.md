## MODIFIED Requirements

### Requirement: Pipeline-first editor layout

coin-ui GP release pipeline editor SHALL present **Composition**, **Pipeline tasks**, **Containerfiles catalog**, and **Parameters** as primary editing sections on GP release detail for draft releases. Editor MUST NOT show separate cards for Build targets, Deliverables, or legacy v2 Containerfile artifacts. Editor MUST NOT live under Platform → Build stacks routes.

#### Scenario: Open pipeline on GP draft release detail

- **WHEN** publisher opens GP draft release detail for `go-app@1.0.0-snapshot.1`
- **THEN** UI MUST show Composition, Pipeline tasks, Containerfiles, and Parameters sections on the release detail page
- **AND** MUST NOT require navigation to `/platform/build-stacks`

#### Scenario: Preview panel on release detail

- **WHEN** publisher edits embedded pipeline on GP draft release detail
- **THEN** resolved manifest preview MUST appear alongside the editor
- **AND** preview MUST show `containerfiles` catalog and `pipeline.tasks`

### Requirement: Task and step editor

Pipeline editor on GP release detail SHALL provide task cards with ordered steps for kinds `coin`, `containerfile`, and `sh`. Editor MUST support `runAfter` selection between tasks. Editor MUST NOT embed managed Containerfile body inside step cards.

#### Scenario: Containerfile step selects catalog ref

- **WHEN** publisher adds or edits `kind: containerfile` step on GP release pipeline editor
- **THEN** UI MUST offer dropdown of `containerfiles[].id` from the same GP draft
- **AND** MUST NOT show inline `containerfile.body` textarea on the step card

#### Scenario: Coin build step references catalog

- **WHEN** publisher adds `kind: coin` step with `action: build` and buildkit engine
- **THEN** UI MUST offer `containerfileRef` selecting from catalog entries

#### Scenario: Publish step selects build task

- **WHEN** publisher adds `kind: coin` publish step
- **THEN** UI MUST offer selection of existing task ids with build output from prior tasks

### Requirement: Parameters editor rules

Parameters section SHALL show `allowedValues` only for `type: enum` and `required` checkbox for all types.

#### Scenario: Enum allowed values input

- **WHEN** publisher edits enum allowed values
- **THEN** UI MUST parse comma-separated values on blur without losing comma during typing

### Requirement: Containerfiles catalog editor

GP release pipeline editor SHALL provide a dedicated Containerfiles section to create and edit catalog entries (`kind: managed` with body editor, `kind: project` with path field).

#### Scenario: Add managed catalog entry

- **WHEN** publisher adds managed containerfile `app` in catalog section
- **THEN** UI MUST show body textarea for the catalog entry
- **AND** pipeline steps MUST reference `app` by ref only

#### Scenario: Add project catalog entry

- **WHEN** publisher adds project containerfile with path `Dockerfile`
- **THEN** UI MUST collect `path` without body field

## REMOVED Requirements

### Requirement: Inline step editor by action type

**Reason**: Superseded by v4 task/step editor with kinds `coin`, `containerfile`, `sh`.

**Migration**: UI maps former run/build/publish forms to coin and containerfile step editors.

### Requirement: Short hash identifiers for stages and build outputs

**Reason**: Replaced by semantic `task.id` and catalog entry ids.

**Migration**: UI generates semantic ids on create; v3 drafts migrated on open/save.
