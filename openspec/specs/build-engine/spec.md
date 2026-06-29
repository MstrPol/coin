# build-engine Specification

## Purpose

Build engine contract: `manifest.build.engine` dispatch in coin-executor, single coin-agent pod, typed pipeline stages. Operational details: `docs/adr/coin-ci-runtime.md`.
## Requirements
### Requirement: Build engine in manifest

Manifest SHALL include `build.engine` with value `buildkit`, `buildpack`, or `dockerfile`.

#### Scenario: Resolve emits build engine

- **WHEN** product resolves GP with gp-content defining build.engine
- **THEN** manifest MUST contain matching `build.engine` field

### Requirement: Single agent runtime image

Jenkins pod SHALL use single container from manifest `runtime.image` (coin-agent).

#### Scenario: No dual container pod

- **WHEN** coin-lib prepares K8s pod template
- **THEN** pod MUST NOT include separate jnlp and stack containers

### Requirement: Typed pipeline stages

Pipeline stages SHALL be typed executor stage names without shell script references.

#### Scenario: Stage execution via executor

- **WHEN** Jenkins runs Test stage
- **THEN** coin-lib MUST invoke coin-executor run --stage test without GP shell scripts

### Requirement: Build engine E2E

Local pilot SHALL pass E2E for all three engines. Project documentation SHALL map each engine to its sample GP (`go-app`, `go-app-bp`, `go-app-df`) and SHALL cross-link `docs/adr/coin-ci-runtime.md` for runtime implementation details.

#### Scenario: e2e-build-engines

- **WHEN** running make e2e-build-engines
- **THEN** buildkit, buildpack, and dockerfile sample jobs MUST succeed

#### Scenario: Documentation engine matrix

- **WHEN** a reader consults build engine documentation
- **THEN** docs MUST list all three engines with sample GP names
- **AND** MUST link to `docs/adr/coin-ci-runtime.md` for bootstrap and implementation notes

### Requirement: Documented pilot build implementation

Project documentation SHALL state that on local pilot arm64 container builds for engines `buildkit` and `dockerfile` are executed via podman when the podman socket is available, without starting `buildkitd` in bootstrap.

#### Scenario: Agent bootstrap documentation

- **WHEN** documentation describes coin-agent bootstrap on local pilot
- **THEN** it MUST list `podman system service` as a required bootstrap step
- **AND** it MUST NOT require `buildkitd` startup on arm64 pilot

