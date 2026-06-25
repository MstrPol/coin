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
