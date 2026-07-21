## ADDED Requirements

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
