## ADDED Requirements

### Requirement: Composition documentation cross-links

GP composition documentation in `docs/architecture.md` and `docs/control-plane.md` SHALL remain consistent with the three-pin composition requirements in this specification and SHALL cross-link `docs/adr/coin-ci-runtime.md`.

#### Scenario: Architecture references composition spec

- **WHEN** `docs/architecture.md` describes GP composition slots
- **THEN** it MUST list exactly `agent`, `gp-content`, and `branching-model`
- **AND** MUST link to `docs/adr/coin-ci-runtime.md` for runtime slot materialization
