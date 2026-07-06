# build-engine Specification

## Purpose

Build engine contract: `manifest.build.engine` dispatch in coin-executor, single coin-agent pod, typed pipeline stages. Operational details: `docs/adr/coin-ci-runtime.md`.
## Requirements
### Requirement: Build engine in manifest

Manifest SHALL derive build engine dispatch from inline pipeline step `run.engine` and `build.engine` fields. Manifest MUST NOT require top-level `build.engine` or `build.targets` catalog for pipeline-inline stacks.

#### Scenario: Resolve emits per-step engines

- **WHEN** product resolves GP with pipeline-inline gp-content containing buildkit and dockerfile steps
- **THEN** manifest pipeline steps MUST retain per-step `engine` values for executor dispatch
- **AND** manifest MUST NOT require top-level `build.engine` as sole dispatch source

### Requirement: Single agent runtime image

Jenkins pod SHALL use single container from manifest `runtime.image` (coin-agent).

#### Scenario: No dual container pod

- **WHEN** coin-lib prepares K8s pod template
- **THEN** pod MUST NOT include separate jnlp and stack containers

### Requirement: Typed pipeline stages

Pipeline stages SHALL remain typed executor stage names. Each stage executes inline steps (`run`, `build`, `publish`) without GP shell scripts or Jenkins Shared Library business logic.

#### Scenario: Stage execution via inline steps

- **WHEN** Jenkins runs Test stage from manifest with inline `run` step
- **THEN** coin-executor MUST dispatch the run step using inline engine config
- **AND** coin-lib MUST invoke executor without interpreting build logic in Groovy

### Requirement: Inline step dispatch in executor

coin-executor SHALL execute pipeline-inline steps directly from manifest without requiring separate manifest `build.targets` or `deliverables` sections.

#### Scenario: Execute run step

- **WHEN** executor runs stage step with `action: run`, `run.engine: buildkit` and step-local `containerfile.contentRef`
- **THEN** executor MUST materialize containerfile from that step ref and run the configured target
- **AND** MUST NOT require top-level `artifacts.containerfiles` catalog

#### Scenario: Execute build then publish

- **WHEN** stage contains `build` step with `build.id: app` followed later by `publish` with `buildStepId: app`
- **THEN** executor MUST build output during build step
- **AND** publish step MUST publish the output associated with `build.id: app`

### Requirement: Build engine E2E

Local pilot SHALL pass E2E for both engines. Project documentation SHALL map `buildkit` to GP `go-app` and BYO `dockerfile` to GP `go-app-docker`, and SHALL cross-link `docs/adr/coin-ci-runtime.md` for runtime implementation details.

#### Scenario: e2e-build-engines

- **WHEN** running make e2e-build-engines
- **THEN** buildkit and BYO dockerfile sample jobs MUST succeed

#### Scenario: Documentation engine matrix

- **WHEN** a reader consults build engine documentation
- **THEN** docs MUST list two engines with sample GP names
- **AND** MUST NOT document buildpack as supported engine
- **AND** MUST link to `docs/adr/coin-ci-runtime.md` for bootstrap and implementation notes

### Requirement: BYO dockerfile engine

For `build.engine` `dockerfile`, coin-executor SHALL build from Dockerfile path in product workspace per GP content policy. coin-executor MUST NOT materialize managed Containerfile from gp-content package for this engine.

#### Scenario: BYO build from checkout

- **WHEN** executor runs build stage for dockerfile engine with `build.dockerfile.path` `Dockerfile`
- **THEN** executor MUST use workspace `Dockerfile` as build definition
- **AND** MUST NOT write `.coin/Containerfile` from Nexus content ref

### Requirement: Buildpack engine removed

build.engine value `buildpack` SHALL NOT be accepted in gp-content validate or manifest resolve.

#### Scenario: Reject buildpack engine

- **WHEN** draft content.yaml sets build.engine buildpack
- **THEN** validate-package MUST reject with validation error

### Requirement: Documented pilot build implementation

Project documentation SHALL state that on local pilot arm64 container builds for engines `buildkit` and `dockerfile` are executed via podman when the podman socket is available, without starting `buildkitd` in bootstrap.

#### Scenario: Agent bootstrap documentation

- **WHEN** documentation describes coin-agent bootstrap on local pilot
- **THEN** it MUST list `podman system service` as a required bootstrap step
- **AND** it MUST NOT require `buildkitd` startup on arm64 pilot

