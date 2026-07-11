## MODIFIED Requirements

### Requirement: Three-pin GP draft composition

GP draft and release composition SHALL contain exactly three operator-selected component pins from the platform registry:

1. **`agent`** — CI runtime stack (agent / executor; e.g. `coin-agent`)
2. **`branching-model`** — versioning and publish policy
3. **`gp-content`** — build stack (Containerfile, schema, pipeline)

Platform `lib` SHALL NOT appear in the GP composition map; resolve injects `lib` from `platform_settings.runtime`.

Standalone `executor` SHALL NOT appear in the GP composition map; resolve materializes executor from the selected agent stack.

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

### Requirement: gp-content pinned per GP version not profile

`gp-content` (build stack) SHALL be pinned only in GP **version** composition (`gp_composition` per draft/release), not on the GP profile entity.

#### Scenario: No profile-level gp-content binding

- **WHEN** enabling team views GP profile `xxx` without opening a specific release
- **THEN** coin-api and coin-ui MUST NOT expose a profile → build stack relationship or implied default gp-content for the profile

#### Scenario: Different releases may pin different gp-content

- **WHEN** GP `xxx` has release `1.0.0` with `gp-content/go-app@1.0.0` and draft `2.0.0-snapshot.1` with `gp-content/other-stack@1.0.0`
- **THEN** each version MUST retain its own composition pins independently

### Requirement: Component catalog independence

`agent`, `gp-content`, and `branching-model` components SHALL exist in the platform registry independently of any GP profile.

#### Scenario: gp-content without matching profile

- **WHEN** `gp-content/go-app@1.0.0` is published in the component registry
- **AND** no GP profile named `go-app` exists
- **THEN** the component MUST remain valid and selectable for future GP drafts

#### Scenario: GP profile without matching gp-content

- **WHEN** GP profile `xxx` exists
- **AND** no `gp-content/xxx` component exists
- **THEN** publisher MUST still be able to create a draft by selecting another gp-content component (e.g. `go-app`) from the catalog

### Requirement: Explicit component names in draft API

New draft and publish requests SHALL include explicit component names separate from the GP profile name.

#### Scenario: agentStackName gpContentName branchingModelName required

- **WHEN** publisher creates a draft
- **THEN** the request MUST include `agentStackName`, `gpContentName`, and `branchingModelName`
- **AND** coin-api MUST validate versions against those component names

#### Scenario: Profile name may differ from gp-content name

- **WHEN** publisher creates a draft for GP profile `xxx` with `gpContentName` `go-app`
- **THEN** coin-api MUST store composition referencing `gp-content/go-app` at the pinned version
- **AND** resolve for GP `xxx` MUST materialize manifest using that gp-content bundle and the selected agent stack

## REMOVED Requirements

### Requirement: Two-slot GP draft composition

**Reason**: Operator must choose agent/executor runtime stack per draft alongside branching model and gp-content.

**Migration**: Add `agent` to composition validation and draft wizard; move agent pin from platform runtime to GP draft; platform runtime retains `lib` only.

### Requirement: gp-content name matches profile

**Reason**: Components are authored by platform team before GP binding; profile name is product identity (`coin.goldenPath`), not build stack name.

**Migration**: Replace implicit `gp-content/{profileName}` with `gpContentName` in API and catalog pickers in UI.
