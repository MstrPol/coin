## MODIFIED Requirements

### Requirement: Canonical CI runtime ADR

The project SHALL maintain `docs/adr/coin-ci-runtime.md` as the canonical architecture decision record for Jenkins CI runtime: single `coin-agent` pod, bootstrap steps, two build engines, publish gate layers, and GP **two-pin** composition (`agent`, `branching-model`) with embedded pipeline on the GP release.

The ADR SHALL document runtime agent registry: profile = image name, version = image tag, metadata stores `image` + `digest`, promote is Platform-only.

#### Scenario: Reader finds runtime SoT

- **WHEN** a contributor looks for how Coin CI pods and build engines work
- **THEN** `docs/adr/README.md` MUST list `coin-ci-runtime` as accepted
- **AND** `docs/architecture.md` MUST link to `coin-ci-runtime` for runtime details
- **AND** composition narrative MUST match `gp-release-two-pin` (not a third `gp-content` pin)

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

## ADDED Requirements

### Requirement: Docs index tracks OpenSpec capabilities

`docs/README.md` SHALL present a reading order that points operators to documents consistent with active OpenSpec specs under `openspec/specs/`, and SHALL link `docs/workspace-layout.md` for repository layout.

#### Scenario: Index does not advertise removed package trees

- **WHEN** reader opens `docs/README.md`
- **THEN** the index MUST NOT list `coin-gp-content/` or `coin-branching-models/` as live package sources
- **AND** MUST link publish/authoring how-tos that match Platform + embedded pipeline
