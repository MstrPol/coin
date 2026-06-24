## MODIFIED Requirements

### Requirement: GP detail Build stack tab

The GP release detail page SHALL include a Build stack tab as the primary path to gp-content for that profile.

#### Scenario: View build stack from GP hub

- **WHEN** enabling team opens the Build stack tab on GP hub for profile `go-app`
- **THEN** the UI MUST show gp-content versions for that GP name (primary path)

#### Scenario: Release detail defers to hub

- **WHEN** enabling team views release detail for a GP release
- **THEN** the UI MAY link to the GP hub Build stack tab rather than duplicating full build stack management on release detail
