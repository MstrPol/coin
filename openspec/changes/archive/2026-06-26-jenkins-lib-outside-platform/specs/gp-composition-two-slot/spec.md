## MODIFIED Requirements

### Requirement: Three-pin GP draft composition

GP draft and release composition SHALL contain exactly three operator-selected component pins from the platform registry:

1. **`agent`** — CI runtime stack (agent / executor; e.g. `coin-agent`)
2. **`branching-model`** — versioning and publish policy
3. **`gp-content`** — build stack (Containerfile, schema, pipeline)

Standalone `executor` SHALL NOT appear in the GP composition map; resolve materializes executor from the selected agent stack.

coin-api SHALL NOT inject Jenkins Shared Library (`lib`) from platform settings or any other control-plane source during resolve.

#### Scenario: Create draft with three catalog pins

- **WHEN** publisher creates a draft with `agentStackName` `coin-agent`, `gpContentName` `go-app`, `branchingModelName` `trunk-based`, and matching versions in the composition map
- **THEN** coin-api MUST accept the draft
- **AND** MUST require keys `agent`, `gp-content`, and `branching-model` in composition

#### Scenario: Reject lib in GP draft composition

- **WHEN** publisher attempts to create a draft including `lib` in composition
- **THEN** coin-api MUST reject the request with a validation error

#### Scenario: Reject standalone executor in GP draft composition

- **WHEN** publisher attempts to create a draft including `executor` as a separate composition key
- **THEN** coin-api MUST reject the request with a validation error

#### Scenario: Resolve without lib injection

- **WHEN** resolve runs for a published GP release with three-pin composition
- **THEN** coin-api MUST materialize executor from the pinned agent stack
- **AND** MUST NOT add `lib` to the resolved manifest
