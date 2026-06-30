## ADDED Requirements

### Requirement: Delete agent draft from runtime hub

The runtime agent hub SHALL allow publishers to delete draft versions from the Releases tab and release detail page.

#### Scenario: Delete draft from releases list

- **WHEN** publisher views `/platform/runtime/{profile}/releases`
- **AND** a version row has `status = draft`
- **THEN** the UI MUST offer a «Delete» or «Delete draft» action for that row
- **AND** MUST NOT offer delete for `published` rows

#### Scenario: Delete draft from release detail

- **WHEN** publisher opens release detail for an agent draft at `/platform/runtime/{profile}/releases/{version}`
- **THEN** the UI MUST offer «Delete draft» alongside Promote
- **AND** MUST require confirmation before calling the Admin API
- **AND** after successful delete MUST navigate back to the Releases tab

#### Scenario: Non-publisher cannot delete

- **WHEN** user without publisher role views agent draft releases
- **THEN** the UI MUST NOT show delete actions
