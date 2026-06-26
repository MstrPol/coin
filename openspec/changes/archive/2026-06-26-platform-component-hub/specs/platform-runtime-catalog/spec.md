## MODIFIED Requirements

### Requirement: Runtime catalog page

The coin-ui SHALL provide a Platform → Runtime catalog listing **agent stack profiles** only.

The page MUST NOT display a platform lib pin banner, «Current platform lib pin» heading, link to edit lib in Platform settings, or `lib` / `executor` component rows.

#### Scenario: No lib pin banner on runtime page

- **WHEN** enabling team opens `/platform/runtime`
- **THEN** the UI MUST NOT show text «Current platform lib pin» or any lib version pin block
- **AND** MUST NOT fetch platform settings solely to display a lib pin

#### Scenario: List agent stack profiles

- **WHEN** enabling team opens `/platform/runtime`
- **THEN** the UI MUST show agent component profiles (e.g. `coin-agent`, `coin-agent-arm`) with version summary per profile
- **AND** MUST NOT show type `executor` or `lib` components

#### Scenario: Open agent stack hub from catalog

- **WHEN** enabling team selects an agent profile row
- **THEN** the UI MUST navigate to `/platform/runtime/{name}`

### Requirement: Runtime publish guidance

The runtime catalog and hub SHALL support draft registration and promote while preserving CI-first image publish.

#### Scenario: Show publish runbook hint

- **WHEN** enabling team views runtime catalog or agent hub
- **THEN** the UI MUST show a link or hint to the agent publish runbook for CI image push and API register
- **AND** MUST NOT reference Component Studio

#### Scenario: New agent profile from catalog

- **WHEN** publisher clicks «New profile» on `/platform/runtime`
- **THEN** the UI MUST navigate to `/platform/runtime/new` to create an agent component profile
- **AND** after create MUST land on hub with welcome CTA for first draft

#### Scenario: Draft agent versions visible in hub

- **WHEN** enabling team opens agent stack hub Releases tab
- **THEN** the UI MUST list both `draft` and `published` versions for that profile
- **AND** MUST offer «New draft» on the hub for additional versions
