# runtime-agent-registry Specification

## Purpose
TBD - created by archiving change runtime-agent-registry. Update Purpose after archive.
## Requirements
### Requirement: Agent registry metadata contract

Platform agent component versions SHALL store runtime image pin metadata: `image` (full container image reference for Jenkins pod pull) and `digest` (content-addressable `sha256:` sum).

The component profile name (`components.name`) SHALL equal the image repository name segment used in CI (e.g. profile `coin-agent` → image `.../coin-agent:{version}`).

Component version string SHALL equal the image tag parsed from `metadata.image`. The image tag is the sole source of truth for agent version identity on manual register.

#### Scenario: CI registers draft after push

- **WHEN** Jenkins `publish-agent.sh` completes docker push for version `1.2.0`
- **THEN** coin-api MUST create `agent/{profile}@1.2.0` with `status = draft`
- **AND** metadata MUST include `image` with tag `1.2.0` and `digest` from the push
- **AND** MUST NOT call promote

#### Scenario: Image tag matches version

- **WHEN** publisher promotes agent version `1.2.0`
- **THEN** coin-api MUST verify parsed tag from `metadata.image` equals `1.2.0`
- **AND** MUST reject promote if tag mismatch

#### Scenario: Manual draft version derived from image

- **WHEN** publisher registers agent draft via Platform UI or API with `metadata.image` `nexus:8082/coin-docker/coin-agent:1.2.0`
- **AND** does not rely on a separate version field
- **THEN** coin-api MUST create `agent/coin-agent@1.2.0` with parsed version `1.2.0`
- **AND** MUST validate repository segment `coin-agent` matches profile name

#### Scenario: Reject version field mismatch on agent draft create

- **WHEN** client POSTs agent draft with `version` `9.9.9` and `metadata.image` ending with `:1.2.0`
- **THEN** coin-api MUST reject with HTTP 422

#### Scenario: Reject unparseable image tag

- **WHEN** client POSTs agent draft with `metadata.image` without a tag (no `:` after repository segment) or tag `latest`
- **THEN** coin-api MUST reject with HTTP 422 on `metadata.image`

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

### Requirement: Agent image tag parsing

coin-api SHALL parse agent version from `metadata.image` using: repository segment after last `/`, tag after last `:` in that segment, ignoring optional `@sha256:` digest suffix on the reference.

#### Scenario: Parse version from standard registry ref

- **WHEN** `metadata.image` is `nexus:8082/coin-docker/agent-30-06:1.2.0`
- **THEN** parsed version MUST be `1.2.0`
- **AND** parsed repository name MUST be `agent-30-06`

#### Scenario: Host port does not confuse tag parse

- **WHEN** `metadata.image` is `nexus:8082/coin-docker/coin-agent:2.0.0`
- **THEN** parsed version MUST be `2.0.0`
- **AND** MUST NOT treat `8082` as the image tag

