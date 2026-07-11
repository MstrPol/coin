# branching-model-preview Specification

## Purpose
TBD - created by archiving change branching-model-schema-v2. Update Purpose after archive.
## Requirements
### Requirement: Branching model preview API absent

coin-api SHALL NOT expose a branching model scenario preview endpoint.

Branching model runtime behavior SHALL be evaluated by `coin-executor` during CI from `manifest.branching`, not by `coin-api` during platform authoring.

#### Scenario: Branching preview endpoint is unavailable

- **WHEN** client calls `POST /v1/admin/branching-models/preview`
- **THEN** coin-api MUST return HTTP 404 Not Found

#### Scenario: coin-api does not import executor

- **WHEN** coin-api is built
- **THEN** the `coin-api` Go module MUST NOT require or import `coin.local/coin-executor`

