## MODIFIED Requirements

### Requirement: Three-pin GP draft composition

GP release composition SHALL contain exactly three operator-controlled pins:

1. **`agent`** — CI runtime stack (container image with baked `coin-executor`; e.g. `coin-agent`, `agent-30-06`)
2. **`gp-content`** — build policy, Containerfile, schema
3. **`branching-model`** — versioning and publish policy

Standalone `executor` SHALL NOT appear in the GP composition map, registry, or resolved manifest. The agent pin is the sole runtime source of truth.

#### Scenario: Create GP draft with three pins

- **WHEN** publisher creates GP draft with agent, gp-content, and branching-model versions
- **THEN** coin-api MUST persist exactly three composition rows
- **AND** MUST NOT validate or require `executor/coin-executor@{agentVersion}` in the component registry

#### Scenario: Reject standalone executor in GP draft composition

- **WHEN** publisher attempts to create a draft including `executor` as a separate composition key
- **THEN** coin-api MUST reject with invalid composition error

#### Scenario: Resolve materializes runtime from agent only

- **WHEN** GP release is resolved for CI
- **THEN** coin-api MUST populate `manifest.runtime` from the pinned agent version metadata
- **AND** MUST NOT add `manifest.executor`
- **AND** MUST NOT query component registry for type `executor`
