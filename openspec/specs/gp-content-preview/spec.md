# gp-content-preview Specification

## Purpose

Superseded by GP release pipeline preview API in `gp-embedded-pipeline`. Retained for delta traceability only.

## Requirements

### Requirement: GP content preview superseded

Standalone gp-content preview API SHALL NOT be exposed. Preview SHALL use GP release pipeline preview endpoint per `gp-embedded-pipeline`.

#### Scenario: Use GP release pipeline preview instead

- **WHEN** publisher needs to preview pipeline-inline model before promote
- **THEN** coin-ui MUST call `POST /v1/admin/golden-paths/{name}/versions/{version}/pipeline/preview`
- **AND** MUST NOT call legacy `/v1/admin/gp-content/preview`
