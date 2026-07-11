## MODIFIED Requirements

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

## ADDED Requirements

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
