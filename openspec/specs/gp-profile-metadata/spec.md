# gp-profile-metadata Specification

## Purpose
TBD - created by archiving change gp-profile-metadata-model. Update Purpose after archive.

## Requirements

### Requirement: GP profile metadata fields

The coin-api GP profile SHALL store only identity metadata: required `name` and optional `description`.

#### Scenario: Create profile with description

- **WHEN** publisher creates a GP profile with `name=go-app` and `description=Go applications golden path`
- **THEN** coin-api MUST persist the profile with those fields and MUST NOT require component slot bindings

#### Scenario: Create profile without description

- **WHEN** publisher creates a GP profile with only `name=go-app`
- **THEN** coin-api MUST persist the profile with null or empty description

### Requirement: Profile read without slots

The coin-api profile API SHALL NOT expose composition slot bindings on the profile entity.

#### Scenario: Get profile

- **WHEN** client calls `GET /v1/admin/golden-paths/{name}/profile`
- **THEN** the response MUST include `name` and `description` only (no `slots` array)

### Requirement: GP profile name equals pipeline family identity

GP profile `name` SHALL be the sole identifier for the embedded pipeline family and MUST match `coin.goldenPath` in product repositories for that Golden Path. coin-api and coin-ui MUST NOT support decoupled alias profiles that reuse another profile's pipeline.

#### Scenario: Reject decoupled pipeline alias pattern

- **WHEN** seed or operator attempts to create GP profile `gp-01-07` that reuses pipeline content intended for profile `go-app` without its own embedded pipeline body
- **THEN** bootstrap and documentation MUST NOT treat this as a supported pattern

#### Scenario: Distinct profiles for distinct pipelines

- **WHEN** platform offers `go-app` and `go-app-docker` Golden Paths
- **THEN** each MUST be a separate GP profile with its own embedded pipeline-inline body
- **AND** product samples MUST reference matching `coin.goldenPath` names
