# runtime-documentation Specification

## Purpose

Documentation and ADR requirements for Coin CI runtime (coin-agent, bootstrap, build engines, publish gate, two-pin GP composition with embedded pipeline). Canonical ADR: `docs/adr/coin-ci-runtime.md`.

## Requirements

### Requirement: Canonical CI runtime ADR

The project SHALL maintain `docs/adr/coin-ci-runtime.md` as the canonical architecture decision record for Jenkins CI runtime: single `coin-agent` pod, bootstrap steps, two build engines, publish gate layers, and GP **two-pin** composition (`agent`, `branching-model`) with embedded pipeline on the GP release.

The ADR SHALL document runtime agent registry: profile = image name, version = image tag, metadata stores `image` + `digest`, promote is Platform-only.

#### Scenario: Reader finds runtime SoT

- **WHEN** a contributor looks for how Coin CI pods and build engines work
- **THEN** `docs/adr/README.md` MUST list `coin-ci-runtime` as accepted
- **AND** `docs/architecture.md` MUST link to `coin-ci-runtime` for runtime details
- **AND** composition narrative MUST match `gp-release-two-pin` (not a third `gp-content` pin)

### Requirement: Superseded ADR banners

Superseded ADR files SHALL include a visible superseded banner at the top with replacement pointers.

#### Scenario: Open superseded composition ADR

- **WHEN** a reader opens `docs/adr/gp-composition-four-components.md` or `docs/adr/gp-pipeline-bundle-layer.md`
- **THEN** the file MUST state status superseded and link to current replacements (`coin-ci-runtime`, `jenkins-lib-http-nexus`, `gp-composition-two-slot` narrative)

### Requirement: Top-level docs aligned with two-pin composition

`docs/architecture.md` and `docs/control-plane.md` SHALL describe GP composition as two operator pins (`agent`, `branching-model`) plus embedded pipeline on the GP release, and SHALL NOT describe standalone `executor`, `lib`, or `gp-content` as GP composition keys.

`docs/adr/coin-ci-runtime.md` SHALL state that the agent pin defines the full CI runtime stack (image + baked `coin-executor` binary) and that resolved manifest v1 SHALL NOT include an `executor` section.

#### Scenario: Architecture doc two-pin model

- **WHEN** reader consults architecture documentation for GP composition
- **THEN** it MUST list exactly two GP composition slots (`agent`, `branching-model`)
- **AND** MUST describe pipeline as embedded on GP release (not a composition pin)
- **AND** MUST NOT list `executor` or `gp-content` as GP composition slots
- **AND** it MUST state that `coin-executor` runs from the agent container image, not from a separate registry component

#### Scenario: Manifest schema without executor

- **WHEN** `manifest.schema.json` defines resolved manifest v1
- **THEN** `executor` MUST NOT appear in `required` or `properties`
- **AND** `runtime` MUST remain required with `image` (and optional `digest`)

### Requirement: Publish gate documentation

Documentation examples SHALL describe publish eligibility using Jenkins `params.publish` and `manifest.branching`, not `pipeline.stages[].when: tag` as the primary gate.

#### Scenario: Control plane manifest example

- **WHEN** a reader opens `docs/control-plane.md` manifest pipeline example
- **THEN** the publish stage MUST NOT use `when: tag` as the documented primary publish gate
- **AND** the doc MUST reference branching policy and `params.publish` per `docs/adr/gp-branching-model.md`

### Requirement: Pilot vs corp build implementation

`docs/adr/coin-ci-runtime.md` SHALL document two product engines (`buildkit`, BYO `dockerfile`) and that buildpack is superseded. On local pilot arm64 both engines use podman build when podman socket is available, and corp amd64 targets native buildkitd per roadmap.

#### Scenario: Pilot troubleshooting

- **WHEN** a reader debugs arm64 pilot build failures
- **THEN** `docs/agent-build-model.md` MUST explain podman-first implementation for both engines
- **AND** MUST NOT list buildpack bootstrap steps
- **AND** MUST link to `coin-ci-runtime` for environment matrix

### Requirement: Agent publish runbook documents manual promote

`docs/agent-build-model.md` (or linked how-to) SHALL document the two-step agent release: CI push + draft register, then Platform promote.

The runbook MUST NOT document GOARCH as a platform metadata field.

#### Scenario: Operator publishes new agent

- **WHEN** a reader follows the agent publish runbook
- **THEN** steps MUST include Platform UI promote after CI draft register
- **AND** MUST NOT instruct CI to auto-promote agent versions

### Requirement: Documentation seed path for GP pipeline defaults

Project documentation and ADRs SHALL describe bootstrap GP pipeline defaults as living in `coin-api/internal/gpcontent/seed/`, and SHALL NOT present `coin/coin-gp-content/` as the source of truth for build engine policy or pipeline content.

#### Scenario: Reader finds seed SoT

- **WHEN** a contributor opens `docs/adr/gp-embedded-pipeline.md` or build-engine / architecture docs for where default pipeline content comes from
- **THEN** the docs MUST point to coin-api seed embed (and GP release authoring for changes)
- **AND** MUST NOT require navigating to `coin-gp-content/stacks/` as live SoT

#### Scenario: Architecture inventory without coin-gp-content package repo

- **WHEN** reader consults `docs/architecture.md` or `docs/control-plane.md` component inventory
- **THEN** they MUST NOT list `coin-gp-content` as an active package publisher required for GP resolve
- **AND** MUST describe embedded pipeline on GP release per accepted ADR

### Requirement: Docs index tracks OpenSpec capabilities

`docs/README.md` SHALL present a reading order that points operators to documents consistent with active OpenSpec specs under `openspec/specs/`, and SHALL link `docs/workspace-layout.md` for repository layout.

#### Scenario: Index does not advertise removed package trees

- **WHEN** reader opens `docs/README.md`
- **THEN** the index MUST NOT list `coin-gp-content/` or `coin-branching-models/` as live package sources
- **AND** MUST link publish/authoring how-tos that match Platform + embedded pipeline
