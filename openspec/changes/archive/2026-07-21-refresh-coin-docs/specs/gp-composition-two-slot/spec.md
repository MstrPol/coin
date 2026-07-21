## MODIFIED Requirements

### Requirement: Documentation consistency for two-pin composition

GP composition documentation in `docs/architecture.md`, `docs/control-plane.md`, and `docs/golden-paths.md` SHALL remain consistent with the two-pin composition requirements in this specification and SHALL cross-link `docs/adr/coin-ci-runtime.md` and `docs/adr/gp-embedded-pipeline.md`.

#### Scenario: Golden paths doc matches two-pin

- **WHEN** reader opens `docs/golden-paths.md` for composition rules
- **THEN** the document MUST describe pins `agent` and `branching-model` only
- **AND** MUST state that pipeline-inline lives on the GP release body
- **AND** MUST NOT present `gp-content` as a required composition pin for new releases
