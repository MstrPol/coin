## MODIFIED Requirements

### Requirement: GP hub entity page

The coin-ui SHALL provide a GP hub at `/gp/{name}` as the primary place to manage one Golden Path profile.

#### Scenario: Hub tabs

- **WHEN** enabling team opens a GP hub
- **THEN** the UI MUST offer tabs: Overview, Releases, Policy, and Canary
- **AND** MUST NOT offer a Build stack tab on the profile hub (no profile ↔ gp-content relationship)

#### Scenario: Policy tab content

- **WHEN** enabling team opens the Policy tab for `go-app`
- **THEN** the UI MUST show the same version policy controls as the former `/catalog` page for that GP

#### Scenario: Canary tab content

- **WHEN** enabling team opens the Canary tab for `go-app`
- **THEN** the UI MUST show the same canary rollout controls as the former `/canary` page for that GP

#### Scenario: Releases tab content

- **WHEN** enabling team opens the Releases tab
- **THEN** the UI MUST list releases for that GP only (published and drafts per existing filters)

#### Scenario: Release detail shows version composition

- **WHEN** enabling team opens release detail for GP `xxx` version `1.0.0`
- **THEN** the UI MUST show the composition table for **that version** (agent, gp-content, branching-model pins)
- **AND** agent pin MUST link to `/platform/runtime/{agentName}/releases/{version}`
- **AND** gp-content pin MUST link to `/platform/build-stacks/{name}/releases/{version}` or edit route for draft
- **AND** branching-model pin MUST link to `/platform/branching-models/{name}/releases/{version}` or edit route for draft
- **AND** MUST NOT link to flat catalog URLs or Component Studio

#### Scenario: Draft release detail actions

- **WHEN** publisher opens release detail for a draft
- **THEN** the UI MUST offer promote and delete draft actions
- **AND** MUST NOT offer delete for published releases
