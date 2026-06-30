## MODIFIED Requirements

### Requirement: Canonical CI runtime ADR

The project SHALL maintain `docs/adr/coin-ci-runtime.md` as the canonical architecture decision record for Jenkins CI runtime: single `coin-agent` pod, bootstrap steps, two build engines, publish gate layers, and GP three-pin composition references.

The ADR SHALL document runtime agent registry: profile = image name, version = image tag, metadata stores `image` + `digest`, promote is Platform-only.

#### Scenario: Reader finds runtime SoT

- **WHEN** a contributor looks for how Coin CI pods and build engines work
- **THEN** `docs/adr/README.md` MUST list `coin-ci-runtime` as accepted
- **AND** `docs/architecture.md` MUST link to `coin-ci-runtime` for runtime details

## ADDED Requirements

### Requirement: Agent publish runbook documents manual promote

`docs/agent-build-model.md` (or linked how-to) SHALL document the two-step agent release: CI push + draft register, then Platform promote.

The runbook MUST NOT document GOARCH as a platform metadata field.

#### Scenario: Operator publishes new agent

- **WHEN** a reader follows the agent publish runbook
- **THEN** steps MUST include Platform UI promote after CI draft register
- **AND** MUST NOT instruct CI to auto-promote agent versions
