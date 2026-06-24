## ADDED Requirements

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
