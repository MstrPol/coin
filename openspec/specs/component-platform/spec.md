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

The system SHALL store immutable component packages in Nexus with `package.manifest.json` listing files, sha256, and roles.

#### Scenario: Register version from package

- **WHEN** enabling team publishes a component version via Admin API
- **THEN** the system MUST upload the package to Nexus, compute digest, and store `content_ref` v2 in `component_versions` without duplicating large bodies in PostgreSQL

### Requirement: Generic resolve materializers

The system SHALL resolve GP composition slots through a type registry and materializers, not per-type switch statements.

#### Scenario: Materialize gp-content slot

- **WHEN** composition pins `gp-content/go-app@1.0.2`
- **THEN** resolve MUST materialize `build`, `pipeline`, and related manifest sections from the registered package

### Requirement: Nexus fallback without PostgreSQL

The system SHALL support product CI resolve from Nexus manifest blobs when coin-api PostgreSQL is unavailable.

#### Scenario: Fallback resolve

- **WHEN** coin-api is unreachable and Nexus has a cached manifest blob for the GP release
- **THEN** coin-lib resolve fallback MUST produce a complete manifest for all pinned slots

### Requirement: Promote canary to stable

The system SHALL support promoting canary line to stable after health validation.

#### Scenario: Catalog promote

- **WHEN** platform lead promotes canary after health gate passes
- **THEN** `catalog.latest` MUST point to the former `latest_canary` composition and component versions MUST transition from `canary` to `published` where applicable

