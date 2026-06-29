## ADDED Requirements

### Requirement: Canonical CI runtime ADR

The project SHALL maintain `docs/adr/coin-ci-runtime.md` as the canonical architecture decision record for Jenkins CI runtime: single `coin-agent` pod, bootstrap steps, three build engines, publish gate layers, and GP three-pin composition references.

#### Scenario: Reader finds runtime SoT

- **WHEN** a contributor looks for how Coin CI pods and build engines work
- **THEN** `docs/adr/README.md` MUST list `coin-ci-runtime` as accepted
- **AND** `docs/architecture.md` MUST link to `coin-ci-runtime` for runtime details

### Requirement: Superseded ADR banners

Superseded ADR files SHALL include a visible superseded banner at the top with replacement pointers.

#### Scenario: Open superseded composition ADR

- **WHEN** a reader opens `docs/adr/gp-composition-four-components.md` or `docs/adr/gp-pipeline-bundle-layer.md`
- **THEN** the file MUST state status superseded and link to current replacements (`coin-ci-runtime`, `jenkins-lib-http-nexus`, `gp-composition-two-slot` narrative)

### Requirement: Top-level docs aligned with three-pin composition

`docs/architecture.md` and `docs/control-plane.md` SHALL describe GP composition as three operator pins (`agent`, `gp-content`, `branching-model`) and SHALL NOT describe standalone `executor` or `lib` as GP composition keys.

#### Scenario: Architecture composition table

- **WHEN** a reader opens `docs/architecture.md` GP composition section
- **THEN** it MUST NOT list `executor` as a separate GP composition slot
- **AND** it MUST state executor is derived from the pinned agent stack

### Requirement: Publish gate documentation

Documentation examples SHALL describe publish eligibility using Jenkins `params.publish` and `manifest.branching`, not `pipeline.stages[].when: tag` as the primary gate.

#### Scenario: Control plane manifest example

- **WHEN** a reader opens `docs/control-plane.md` manifest pipeline example
- **THEN** the publish stage MUST NOT use `when: tag` as the documented primary publish gate
- **AND** the doc MUST reference branching policy and `params.publish` per `docs/adr/gp-branching-model.md`

### Requirement: Pilot vs corp build implementation

`docs/adr/coin-ci-runtime.md` SHALL document that on local pilot arm64 the `buildkit` and `dockerfile` engines use podman build while engine names in manifest remain unchanged, and that corp amd64 targets native buildkitd per roadmap.

#### Scenario: Pilot troubleshooting

- **WHEN** a reader debugs arm64 pilot build failures
- **THEN** `docs/agent-build-model.md` MUST explain podman-first implementation and link to `coin-ci-runtime` for environment matrix
