## ADDED Requirements

### Requirement: Two-pin GP release composition

GP release composition SHALL contain exactly two operator-controlled external pins:

1. **`agent`** — CI runtime stack (e.g. `coin-agent`)
2. **`branching-model`** — versioning and publish policy

Embedded pipeline-inline model on the GP release SHALL NOT appear as a composition pin. Standalone `gp-content` and `executor` SHALL NOT appear in composition.

#### Scenario: Create GP draft with two pins

- **WHEN** publisher creates GP draft with agent and branching-model versions
- **THEN** coin-api MUST persist exactly two composition rows
- **AND** MUST NOT require or accept `gp-content` composition key

#### Scenario: Reject gp-content in composition

- **WHEN** publisher attempts to create or update draft including `gp-content` as composition type
- **THEN** coin-api MUST reject with invalid composition error

### Requirement: Draft API without gpContentName

New draft and publish requests SHALL include `agentStackName` and `branchingModelName` only. Requests MUST NOT include `gpContentName`.

#### Scenario: Create draft without gpContentName

- **WHEN** publisher creates GP draft for profile `go-app`
- **THEN** request MUST include `agentStackName` and `branchingModelName`
- **AND** coin-api MUST NOT accept `gpContentName` field

#### Scenario: Profile name is pipeline identity

- **WHEN** GP profile is named `go-app`
- **THEN** embedded pipeline for releases of that profile MUST be authored for `go-app` family
- **AND** coin-api MUST NOT support pinning another profile's pipeline via gp-content reuse

### Requirement: Promote gate without gp-content pin

GP promote SHALL validate embedded pipeline is valid and both composition pins are `published`. Promote MUST NOT check gp-content component status.

#### Scenario: Promote blocked by draft branching pin

- **WHEN** publisher promotes GP draft with `branching-model` pin in `draft` status
- **THEN** coin-api MUST reject promote with HTTP 409 Conflict

#### Scenario: Promote succeeds with valid pipeline and published pins

- **WHEN** embedded pipeline is valid and agent and branching-model pins are `published`
- **THEN** coin-api MUST complete GP promote

### Requirement: Resolve merges two pins and embedded pipeline

Resolved manifest SHALL materialize `runtime` from agent pin, `branching` from branching-model pin, and `parameters`, `validateSchema`, `pipeline` from embedded GP release pipeline body.

#### Scenario: Resolve published GP release

- **WHEN** GP release `go-app@1.0.0` is resolved for CI
- **THEN** manifest MUST contain `goldenPath.name` and `goldenPath.version`
- **AND** MUST contain pipeline sections from embedded release body
- **AND** MUST NOT reference `gp-content` component in resolve metadata

## REMOVED Requirements

### Requirement: Three-pin GP draft composition

**Reason**: Pipeline is intrinsic to GP release; external gp-content pin removed.

**Migration**: Use two-pin composition plus embedded pipeline on GP release draft. Remove `gpContentName` from all clients.

### Requirement: gp-content pinned per GP version not profile

**Reason**: gp-content component type eliminated; pipeline is GP release payload.

**Migration**: Edit pipeline on GP release detail; no Platform build stacks catalog.

### Requirement: Component catalog independence for gp-content

**Reason**: gp-content no longer exists as registry type.

**Migration**: Create new GP profile per pipeline family (`go-app`, `go-app-docker`).

### Requirement: Explicit component names include gpContentName

**Reason**: gpContentName removed from draft API.

**Migration**: Use `agentStackName` and `branchingModelName` only.

### Requirement: Accept draft gp-content in GP draft

**Reason**: No gp-content pin; pipeline edited on GP release.

**Migration**: Edit embedded pipeline on GP draft directly.

### Requirement: GP draft on draft component pins with promote gate for gp-content

**Reason**: gp-content component lifecycle removed.

**Migration**: Promote gate checks embedded pipeline validity plus agent and branching-model published status.
