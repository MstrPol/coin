# component-platform Specification

## Purpose

Platform components: lifecycle, package model, resolve materializers, Nexus fallback, canary promote.
## Requirements
### Requirement: Component lifecycle states

The system SHALL support component version states: `draft`, `canary`, and `published`.

#### Scenario: Draft not visible to product resolve

- **WHEN** a component version is in `draft` state
- **THEN** resolve for stable and canary channels MUST NOT return that version unless explicitly pinned in a draft GP preview (admin-only)

#### Scenario: Canary visible to pilot projects

- **WHEN** a component version is in `canary` state and GP uses `latest_canary` with that pin
- **THEN** resolve with canary channel MUST return the version for projects with `canary_mode=canary` or canary audience policy

#### Scenario: Published visible to all

- **WHEN** a component version is `published`
- **THEN** resolve on stable channel MUST return it when pinned in GP composition

### Requirement: Component Package Model

The system SHALL store immutable component packages in Nexus with `package.manifest.json` listing files, sha256, and roles for **published** component versions only.

#### Scenario: Register canary version without Nexus

- **WHEN** enabling team registers a `branching-model` draft or canary version via Admin API
- **THEN** the system MUST store artifact bodies and content_ref manifest subset in PostgreSQL and MUST NOT write package files to Nexus

#### Scenario: Publish stable version to Nexus

- **WHEN** enabling team promotes a `branching-model` version from `canary` to `published`
- **THEN** the system MUST upload the package to Nexus, compute digest, and store full `content_ref` v2 including `package.url` and `package.sha256`

### Requirement: Generic resolve materializers

The system SHALL resolve GP composition slots through a type registry and materializers, not per-type switch statements.

#### Scenario: Materialize gp-content slot

- **WHEN** composition pins `gp-content/go-app@1.0.2`
- **THEN** resolve MUST materialize `build`, `pipeline`, and related manifest sections from the registered package

### Requirement: Nexus fallback without PostgreSQL

The system SHALL support product CI resolve from Nexus manifest blobs when coin-api PostgreSQL is unavailable for **published** component versions.

#### Scenario: Fallback resolve

- **WHEN** coin-api is unreachable and Nexus has a cached manifest blob for the GP release with published component packages
- **THEN** coin-lib resolve fallback MUST produce a complete manifest for all pinned slots that have Nexus packages

#### Scenario: Canary requires control plane

- **WHEN** product CI resolves a canary channel pin that references a `branching-model` version without Nexus package
- **THEN** resolve MUST require coin-api PostgreSQL and MUST materialize branching rules from PG content_ref and artifact bodies

### Requirement: Promote canary to stable

The system SHALL support promoting canary line to stable after health validation.

#### Scenario: Catalog promote

- **WHEN** platform lead promotes canary after health gate passes
- **THEN** `catalog.latest` MUST point to the former `latest_canary` composition and component versions MUST transition from `canary` to `published` where applicable

