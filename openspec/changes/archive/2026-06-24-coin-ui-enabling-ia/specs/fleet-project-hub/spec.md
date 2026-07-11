## ADDED Requirements

### Requirement: Fleet navigation grouping

The coin-ui SHALL group project-centric views under a Fleet sidebar section.

#### Scenario: Fleet nav entries

- **WHEN** user views the Fleet group in the sidebar
- **THEN** the UI MUST include Projects and Build reports entries

#### Scenario: Build reports reachable from Fleet

- **WHEN** user clicks Build reports in Fleet
- **THEN** the UI MUST navigate to `/build-reports`

### Requirement: Projects page fleet context

The Projects list SHALL emphasize operational context for enabling team.

#### Scenario: Show GP pin on project row

- **WHEN** enabling team opens the Projects page
- **THEN** each project row MUST show pinned GP profile and version when available from existing API fields
