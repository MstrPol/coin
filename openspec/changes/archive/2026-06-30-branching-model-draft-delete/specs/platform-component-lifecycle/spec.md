## ADDED Requirements

### Requirement: Delete branching-model draft via Admin API

coin-api SHALL accept delete draft requests for `branching-model` component versions through the generic component delete endpoint.

#### Scenario: Delete branching-model draft succeeds

- **WHEN** publisher calls `DELETE /v1/admin/components/branching-model/trunk-based/versions/2.0.0-draft`
- **AND** the version has `status = draft`
- **THEN** coin-api MUST return HTTP 204 No Content
- **AND** MUST remove the `component_versions` row
- **AND** MUST cascade-delete `component_artifact_bodies` rows for that version (e.g. `model.yaml`)
- **AND** MUST write audit log action `delete_component_draft`
- **AND** MUST NOT modify Nexus

#### Scenario: Reject delete published branching-model

- **WHEN** publisher attempts to delete `branching-model/trunk-based@1.0.0` with `status = published`
- **THEN** coin-api MUST return HTTP 409 Conflict
