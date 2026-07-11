## ADDED Requirements

### Requirement: Containerfile catalog on GP release body

GP release pipeline body SHALL include `containerfiles[]` as a named catalog scoped to that GP release version. Catalog entries MUST be referenced by pipeline steps; duplicate managed bodies MUST be deduplicated into one catalog entry per unique content.

#### Scenario: Multiple catalog entries on one GP

- **WHEN** publisher defines `containerfiles` with ids `app` and `liquibase` on GP draft
- **THEN** coin-api MUST persist both entries in the GP release pipeline body
- **AND** steps MUST reference entries by `ref` only

#### Scenario: Reject orphan catalog entry

- **WHEN** catalog entry is not referenced by any pipeline step and is not marked `reserved: true` in seed templates
- **THEN** validate MUST warn or reject per pilot policy

### Requirement: Managed containerfile catalog entry

Catalog entry with `kind: managed` SHALL include non-empty `body` with Containerfile content. Author model MUST NOT store managed body inline on pipeline steps.

#### Scenario: Managed entry materialization

- **WHEN** GP draft promotes with managed catalog entry `app`
- **THEN** manifest MUST include `containerfiles[].contentRef` and `digest` for `app`
- **AND** pipeline steps referencing `app` MUST resolve content from manifest catalog

### Requirement: Project containerfile catalog entry

Catalog entry with `kind: project` SHALL include workspace-relative `path` to a Containerfile or Dockerfile in the product repo. Executor MUST NOT materialize managed body from Nexus for project entries.

#### Scenario: BYO project path entry

- **WHEN** catalog entry has `kind: project` and `path: Dockerfile`
- **THEN** validate MUST accept entry without `body`
- **AND** executor MUST use workspace file at `path` when step references the entry

### Requirement: Catalog entry id format

`containerfiles[].id` SHALL match `^[a-z][a-z0-9-]{1,31}$` and MUST be unique within the GP release pipeline body.

#### Scenario: Reject duplicate catalog id

- **WHEN** two catalog entries share the same `id`
- **THEN** validate MUST reject the draft
