## MODIFIED Requirements

### Requirement: Release detail under platform hub

Platform component release detail SHALL live under the platform hub URL hierarchy.

For `agent` component releases, the page SHALL display agent runtime metadata (`image`, `digest`) and MUST NOT display a derived `executor/coin-executor` pin.

#### Scenario: Agent release detail runtime metadata only

- **WHEN** user opens `/platform/runtime/coin-agent/releases/1.0.0`
- **THEN** the UI MUST show `image` and `digest` from agent metadata
- **AND** MUST NOT show a derived executor pin section
- **AND** API response MUST NOT include `derivedExecutorPin`
