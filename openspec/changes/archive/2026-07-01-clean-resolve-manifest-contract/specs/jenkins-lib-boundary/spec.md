## ADDED Requirements

### Requirement: Jenkins credentials outside resolved manifest

Jenkins credential IDs SHALL be selected by product/Jenkins configuration and `coin-lib` defaults, not by `coin-api` resolved manifest.

Resolved manifest SHALL NOT contain a top-level `credentials` section or any Jenkins-instance credential ID such as `nexus-docker`.

`coin-lib` SHALL NOT merge `manifest.credentials` into the effective project configuration. Docker registry credential binding SHALL use `jenkins.credentials.docker` from product config or the existing `coin-lib` default chain.

#### Scenario: Resolve manifest excludes Jenkins credential IDs

- **WHEN** product CI resolves a GP manifest
- **THEN** the returned JSON document MUST NOT contain top-level key `credentials`
- **AND** MUST NOT contain Jenkins credential ID values used for local registry binding

#### Scenario: coin-lib binds Docker credentials from project config

- **WHEN** product `.coin/config.yaml` contains `jenkins.credentials.docker: nexus-docker`
- **AND** the resolved manifest has no `credentials` key
- **THEN** `coin-lib` MUST bind Docker registry credentials using the product config value

#### Scenario: coin-lib default remains local pilot fallback

- **WHEN** product config omits an optional credential value that `coin-lib` supports through defaults
- **THEN** `coin-lib` MAY use its own defaults or Jenkins environment configuration
- **AND** `coin-api` MUST NOT provide that fallback through manifest resolve
