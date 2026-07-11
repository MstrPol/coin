## MODIFIED Requirements

### Requirement: Release detail under platform hub

Platform component release detail SHALL live under the platform hub URL hierarchy.

#### Scenario: Agent release detail shows derived executor

- **WHEN** enabling team opens release detail for agent `coin-agent` version `1.0.0`
- **THEN** the UI MUST show agent metadata (`image`, `digest`)
- **AND** MUST show read-only derived pin `executor/coin-executor@1.0.0`
- **AND** MUST NOT list executor as a separate release in the Releases tab
- **AND** MUST NOT show GOARCH or architecture fields

#### Scenario: Draft release detail actions

- **WHEN** publisher opens release detail for a draft platform component version
- **THEN** the UI MUST offer promote and delete draft actions appropriate to the component type
- **AND** for `agent` drafts MUST wire delete draft to `DELETE /v1/admin/components/agent/{name}/versions/{version}`
- **AND** for `branching-model` drafts MUST wire delete draft to `DELETE /v1/admin/components/branching-model/{name}/versions/{version}` on the Releases tab and in the draft editor lifecycle panel
- **AND** MUST NOT offer delete for published releases
