## ADDED Requirements

### Requirement: Documented pilot build implementation

Project documentation SHALL state that on local pilot arm64 container builds for engines `buildkit` and `dockerfile` are executed via podman when the podman socket is available, without starting `buildkitd` in bootstrap.

#### Scenario: Agent bootstrap documentation

- **WHEN** documentation describes coin-agent bootstrap on local pilot
- **THEN** it MUST list `podman system service` as a required bootstrap step
- **AND** it MUST NOT require `buildkitd` startup on arm64 pilot

## MODIFIED Requirements

### Requirement: Build engine E2E

Local pilot SHALL pass E2E for all three engines. Project documentation SHALL map each engine to its sample GP (`go-app`, `go-app-bp`, `go-app-df`) and SHALL cross-link `docs/adr/coin-ci-runtime.md` for runtime implementation details.

#### Scenario: e2e-build-engines

- **WHEN** running make e2e-build-engines
- **THEN** buildkit, buildpack, and dockerfile sample jobs MUST succeed

#### Scenario: Documentation engine matrix

- **WHEN** a reader consults build engine documentation
- **THEN** docs MUST list all three engines with sample GP names
- **AND** MUST link to `docs/adr/coin-ci-runtime.md` for bootstrap and implementation notes
