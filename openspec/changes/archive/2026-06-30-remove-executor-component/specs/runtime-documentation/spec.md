## MODIFIED Requirements

### Requirement: Top-level docs aligned with three-pin composition

`docs/architecture.md` and `docs/control-plane.md` SHALL describe GP composition as three operator pins (`agent`, `gp-content`, `branching-model`) and SHALL NOT describe standalone `executor` or `lib` as GP composition keys.

`docs/adr/coin-ci-runtime.md` SHALL state that the agent pin defines the full CI runtime stack (image + baked `coin-executor` binary) and that resolved manifest v1 SHALL NOT include an `executor` section.

#### Scenario: Architecture doc three-pin model

- **WHEN** reader consults architecture documentation for GP composition
- **THEN** it MUST list exactly three GP composition slots
- **THEN** it MUST NOT list `executor` as a separate GP composition slot or manifest section
- **AND** it MUST state that `coin-executor` runs from the agent container image, not from a separate registry component

#### Scenario: Manifest schema without executor

- **WHEN** `manifest.schema.json` defines resolved manifest v1
- **THEN** `executor` MUST NOT appear in `required` or `properties`
- **AND** `runtime` MUST remain required with `image` (and optional `digest`)
