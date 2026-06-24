# fleet-project-hub Specification

## Purpose

Fleet navigation grouping and project list operational context in coin-ui.

## Requirements

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

### Requirement: Projects paginated table

The Projects page SHALL display one page of results at a time using server-side pagination.

#### Scenario: Navigate pages

- **WHEN** enabling team clicks next page on Projects
- **THEN** the UI MUST request the next `offset` from the API and MUST NOT load all projects into memory

#### Scenario: Show total count

- **WHEN** projects list loads
- **THEN** the UI MUST display total matching project count from the API `total` field

### Requirement: Projects pagination in URL

Pagination state on Projects SHALL be reflected in URL search parameters alongside existing filters.

#### Scenario: Bookmark filtered page

- **WHEN** user opens a URL with `goldenPath`, `page`, and `pageSize` on Projects
- **THEN** the UI MUST restore filters and the same page of results
