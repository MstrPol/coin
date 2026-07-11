## MODIFIED Requirements

### Requirement: Build engine in manifest

Manifest SHALL include `build.engine` with value `buildkit` or `dockerfile`.

#### Scenario: Resolve emits build engine

- **WHEN** product resolves GP with gp-content defining build.engine
- **THEN** manifest MUST contain matching `build.engine` field

### Requirement: Build engine E2E

Local pilot SHALL pass E2E for both engines. Project documentation SHALL map `buildkit` to GP `go-app` and BYO `dockerfile` to GP `go-app-docker`.

#### Scenario: e2e-build-engines

- **WHEN** running make e2e-build-engines
- **THEN** buildkit and BYO dockerfile sample jobs MUST succeed

#### Scenario: Documentation engine matrix

- **WHEN** a reader consults build engine documentation
- **THEN** docs MUST list two engines with sample GP names
- **AND** MUST NOT document buildpack as supported engine

## ADDED Requirements

### Requirement: BYO dockerfile engine

For `build.engine` `dockerfile`, coin-executor SHALL build from Dockerfile path in product workspace per GP content policy. coin-executor MUST NOT materialize managed Containerfile from gp-content package for this engine.

#### Scenario: BYO build from checkout

- **WHEN** executor runs build stage for dockerfile engine with `build.dockerfile.path` `Dockerfile`
- **THEN** executor MUST use workspace `Dockerfile` as build definition
- **AND** MUST NOT write `.coin/Containerfile` from Nexus content ref

### Requirement: Artifact deliverable buildkit only

GP content with `build.engine` `dockerfile` SHALL NOT declare `artifact` in `capabilities.deliverables`. validate-package MUST reject artifact deliverable for BYO engine.

#### Scenario: Reject artifact on BYO GP

- **WHEN** draft gp-content has engine dockerfile and deliverables include artifact
- **THEN** validate-package MUST fail validation

### Requirement: Buildpack engine removed

build.engine value `buildpack` SHALL NOT be accepted in gp-content validate or manifest resolve.

#### Scenario: Reject buildpack engine

- **WHEN** draft content.yaml sets build.engine buildpack
- **THEN** validate-package MUST reject with validation error
