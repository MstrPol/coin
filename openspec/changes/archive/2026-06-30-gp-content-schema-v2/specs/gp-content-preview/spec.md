## ADDED Requirements

### Requirement: GP content preview API

coin-api SHALL expose `POST /v1/admin/gp-content/preview` accepting draft gp-content manifest subset or `content.yaml` body and returning resolved `build` and `pipeline` fragments plus validation issues and engine-specific warnings.

#### Scenario: Preview buildkit draft

- **WHEN** publisher posts valid buildkit `content.yaml` v2 draft
- **THEN** coin-api MUST return resolved `build.engine` `buildkit` with targets and cacheRef template resolution hints
- **AND** MUST include `artifacts.containerfile` when engine is buildkit

#### Scenario: Preview BYO dockerfile draft

- **WHEN** publisher posts valid dockerfile engine v2 draft with `build.dockerfile.path`
- **THEN** coin-api MUST return resolved `build.dockerfile` without managed `containerfile` content ref
- **AND** MUST warn if `capabilities.deliverables` includes `artifact`

#### Scenario: Reject v1 content

- **WHEN** publisher posts content without `schemaVersion: 2`
- **THEN** preview MUST return validation error
