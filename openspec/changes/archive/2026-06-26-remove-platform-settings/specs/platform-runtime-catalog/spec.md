## MODIFIED Requirements

### Requirement: Runtime catalog page

The coin-ui SHALL provide a Platform → Runtime catalog listing **agent stack profiles** only.

The page MUST NOT display a platform lib pin banner, «Current platform lib pin» heading, or `lib` / `executor` component rows.

#### Scenario: No lib pin banner on runtime page

- **WHEN** enabling team opens `/platform/runtime`
- **THEN** the UI MUST NOT show text «Current platform lib pin» or any lib version pin block

#### Scenario: List agent stack profiles

- **WHEN** enabling team opens `/platform/runtime`
- **THEN** the UI MUST show agent component profiles (e.g. `coin-agent`, `coin-agent-arm`) with version summary per profile
- **AND** MUST NOT show type `executor` or `lib` components

#### Scenario: Open agent stack hub from catalog

- **WHEN** enabling team selects an agent profile row
- **THEN** the UI MUST navigate to `/platform/runtime/{name}`
