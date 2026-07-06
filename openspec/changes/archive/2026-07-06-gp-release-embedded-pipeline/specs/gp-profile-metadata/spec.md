## ADDED Requirements

### Requirement: GP profile name equals pipeline family identity

GP profile `name` SHALL be the sole identifier for the embedded pipeline family and MUST match `coin.goldenPath` in product repositories for that Golden Path. coin-api and coin-ui MUST NOT support decoupled alias profiles that reuse another profile's pipeline.

#### Scenario: Reject decoupled pipeline alias pattern

- **WHEN** seed or operator attempts to create GP profile `gp-01-07` that reuses pipeline content intended for profile `go-app` without its own embedded pipeline body
- **THEN** bootstrap and documentation MUST NOT treat this as a supported pattern

#### Scenario: Distinct profiles for distinct pipelines

- **WHEN** platform offers `go-app` and `go-app-docker` Golden Paths
- **THEN** each MUST be a separate GP profile with its own embedded pipeline-inline body
- **AND** product samples MUST reference matching `coin.goldenPath` names
