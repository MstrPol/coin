## MODIFIED Requirements

### Requirement: Component Package Model

The system SHALL store immutable component packages in Nexus with `package.manifest.json` listing files, sha256, and roles for **published** component versions only.

#### Scenario: Register canary version without Nexus

- **WHEN** enabling team registers a `branching-model` draft or canary version via Admin API
- **THEN** the system MUST store artifact bodies and content_ref manifest subset in PostgreSQL and MUST NOT write package files to Nexus

#### Scenario: Publish stable version to Nexus

- **WHEN** enabling team promotes a `branching-model` version from `canary` to `published`
- **THEN** the system MUST upload the package to Nexus, compute digest, and store full `content_ref` v2 including `package.url` and `package.sha256`

### Requirement: Nexus fallback without PostgreSQL

The system SHALL support product CI resolve from Nexus manifest blobs when coin-api PostgreSQL is unavailable for **published** component versions.

#### Scenario: Fallback resolve

- **WHEN** coin-api is unreachable and Nexus has a cached manifest blob for the GP release with published component packages
- **THEN** coin-lib resolve fallback MUST produce a complete manifest for all pinned slots that have Nexus packages

#### Scenario: Canary requires control plane

- **WHEN** product CI resolves a canary channel pin that references a `branching-model` version without Nexus package
- **THEN** resolve MUST require coin-api PostgreSQL and MUST materialize branching rules from PG content_ref and artifact bodies
