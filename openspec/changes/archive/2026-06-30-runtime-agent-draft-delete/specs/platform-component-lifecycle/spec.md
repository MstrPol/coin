## ADDED Requirements

### Requirement: Admin API delete component draft

coin-api SHALL expose `DELETE /v1/admin/components/{type}/{name}/versions/{version}` to remove a component version with `status = draft`.

The endpoint MUST apply to platform component types with draft lifecycle: `agent`, `gp-content`, and `branching-model`.

The endpoint MUST NOT delete versions with `status = published`.

#### Scenario: Delete agent draft succeeds

- **WHEN** publisher calls `DELETE /v1/admin/components/agent/coin-agent/versions/0.1.0-draft`
- **AND** the version has `status = draft`
- **THEN** coin-api MUST return HTTP 204 No Content
- **AND** MUST remove the `component_versions` row
- **AND** MUST cascade-delete any `component_artifact_bodies` rows for that version
- **AND** MUST write audit log action `delete_component_draft`
- **AND** MUST NOT modify Nexus

#### Scenario: Reject delete published version

- **WHEN** publisher attempts to delete `agent/coin-agent@1.0.0` with `status = published`
- **THEN** coin-api MUST return HTTP 409 Conflict

#### Scenario: Delete not found

- **WHEN** publisher deletes a non-existent component version
- **THEN** coin-api MUST return HTTP 404 Not Found
