# runtime-agent-registry Specification

## Purpose
TBD - created by archiving change runtime-agent-registry. Update Purpose after archive.
## Requirements
### Requirement: Agent registry metadata contract

Platform agent component versions SHALL store runtime image pin metadata: `image` (full container image reference for Jenkins pod pull) and `digest` (content-addressable `sha256:` sum).

The component profile name (`components.name`) SHALL equal the image repository name segment used in CI (e.g. profile `coin-agent` → image `.../coin-agent:{version}`).

Component version semver SHALL equal the image tag in `metadata.image` for published pins.

#### Scenario: CI registers draft after push

- **WHEN** Jenkins `publish-agent.sh` completes docker push for version `1.2.0`
- **THEN** coin-api MUST create `agent/{profile}@1.2.0` with `status = draft`
- **AND** metadata MUST include `image` and `digest` from the push
- **AND** MUST NOT call promote

#### Scenario: Image tag matches version

- **WHEN** publisher promotes agent version `1.2.0`
- **THEN** coin-api MUST verify `metadata.image` ends with `:1.2.0`
- **AND** MUST reject promote if tag mismatch

### Requirement: Manual promote gate

Agent component versions SHALL transition `draft` → `published` only via Platform publisher action (Admin API promote invoked from coin-ui).

CI register endpoints MUST NOT auto-promote agent versions.

#### Scenario: CI does not publish agent

- **WHEN** Jenkins registers agent draft after image push
- **THEN** the version MUST remain `status = draft` until a publisher promotes it in Platform UI

#### Scenario: Promote requires digest

- **WHEN** publisher promotes agent draft without `metadata.digest`
- **THEN** coin-api MUST reject with HTTP 422 and validation error on `metadata.digest`

#### Scenario: Promote requires image

- **WHEN** publisher promotes agent draft without `metadata.image`
- **THEN** coin-api MUST reject with HTTP 422 and validation error on `metadata.image`

### Requirement: No architecture field in agent metadata

Agent component metadata MUST NOT include `goarch` or `architecture` fields. Build architecture is implied by the pinned digest and image manifest.

#### Scenario: Reject goarch on write

- **WHEN** client PATCHes agent draft metadata including `goarch`
- **THEN** coin-api MUST ignore or reject the field per API schema
- **AND** MUST NOT persist `goarch` in component_versions.metadata

### Requirement: Executor derive same-version for all agent profiles

coin-api SHALL derive executor pin for GP resolve as `executor/coin-executor@{agentVersion}` for **any** registered `agent` profile name.

coin-api MUST NOT reject executor derive solely because the agent profile name is not `coin-agent`.

Jenkins pod runtime SHALL use only the pinned agent version `metadata.image` and `metadata.digest`; executor derive MUST NOT require additional fields in agent metadata.

#### Scenario: Resolve GP with alternate agent profile

- **WHEN** GP composition pins `agent/coin-agent-arm@1.2.0` with published agent metadata
- **THEN** coin-api MUST resolve manifest successfully
- **AND** `manifest.runtime` MUST use `image` and `digest` from agent metadata
- **AND** `manifest.executor` MUST reference `executor/coin-executor@1.2.0`

#### Scenario: No hardcoded profile switch

- **WHEN** `executorPinForAgentStack` is called for profile `coin-agent-arm` and version `1.2.0`
- **THEN** coin-api MUST return executor pin `executor/coin-executor@1.2.0`
- **AND** MUST NOT return unsupported agent stack error based on profile name alone

#### Scenario: Executor component must exist at derived version

- **WHEN** resolve derives `executor/coin-executor@1.2.0` for agent pin `1.2.0`
- **AND** executor version `1.2.0` is missing or not visible for resolve mode
- **THEN** coin-api MUST fail resolve with a clear error listing the missing executor pin

