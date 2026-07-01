# jenkins-lib-boundary Specification

## Purpose
TBD - created by archiving change jenkins-lib-outside-platform. Update Purpose after archive.
## Requirements
### Requirement: Jenkins lib outside control plane

The coin-api control plane SHALL NOT register, store, or expose Jenkins Shared Library (`coin-lib`) as a platform component.

Jenkins glue version SHALL be managed outside coin-api (Jenkins org configuration, HTTP retriever, Nexus immutable ZIP).

#### Scenario: No lib component type in registry

- **WHEN** client lists platform components or queries admin component API
- **THEN** coin-api MUST NOT return components with type `lib`

#### Scenario: No lib admin registration API

- **WHEN** client POSTs to register a lib component version
- **THEN** coin-api MUST respond with HTTP 404 Not Found (route removed)

### Requirement: coin-lib Nexus-only publish

The coin-lib repository publish path SHALL upload immutable ZIP artifacts to Nexus only and SHALL NOT call coin-api admin registration.

#### Scenario: Publish without coin-api

- **WHEN** `publish-lib.sh` completes successfully for version `1.0.0`
- **THEN** Nexus MUST contain the ZIP at the configured maven-releases path
- **AND** coin-api PostgreSQL MUST NOT contain a `lib` component row for that version

### Requirement: Manifest resolve excludes lib

Resolved GP manifest SHALL NOT include a `lib` section.

coin-api SHALL NOT read or validate a platform lib pin during resolve.

#### Scenario: Resolve manifest shape

- **WHEN** client resolves manifest for GP `go-app@1.0.0`
- **THEN** the JSON document MUST NOT contain top-level key `lib`
- **AND** MUST contain runtime/executor materialization from agent stack and gp-content/branching pins

### Requirement: Product Jenkins bootstrap independent of coin-api lib API

Product CI SHALL load coin-lib via Jenkins Shared Library configuration (`@Library`) without calling coin-api for lib version selection.

#### Scenario: No LibraryVersion endpoint

- **WHEN** client calls `GET /v1/golden-paths/{name}/version`
- **THEN** coin-api MUST respond with HTTP 404 Not Found

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

