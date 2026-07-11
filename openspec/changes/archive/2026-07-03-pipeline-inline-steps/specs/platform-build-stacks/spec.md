## MODIFIED Requirements

### Requirement: GP content schema v2 editor

The build stack editor SHALL edit pipeline-inline `schemaVersion: 3` through Parameters and Pipeline stages only. Containerfile content MUST be edited inline inside buildkit steps.

#### Scenario: Pipeline-first editing

- **WHEN** publisher edits v3 build stack draft
- **THEN** UI MUST provide inline step forms including containerfile body for buildkit steps
- **AND** MUST NOT require separate targets, deliverables or containerfile catalogs

#### Scenario: Save produces v3 model

- **WHEN** publisher saves draft
- **THEN** package MUST use `schemaVersion: 3` without top-level `artifacts.containerfiles`

## REMOVED Requirements

### Requirement: Engine card switches build block

**Reason**: Replaced by pipeline-inline step editor.
**Migration**: Configure engine per step inline.

### Requirement: Engine card buildkit

**Reason**: Buildkit containerfile and targets configured per step.
**Migration**: Use buildkit step with inline `containerfile.body` and `target`.
