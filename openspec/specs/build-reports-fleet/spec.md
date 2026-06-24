# build-reports-fleet Specification

## Purpose

Build reports fleet page in coin-ui: server-side pagination, date filters, URL state.

## Requirements

### Requirement: Build reports paginated table

The Build reports page SHALL display one page of results at a time using server-side pagination.

#### Scenario: Navigate pages

- **WHEN** enabling team clicks next page on Build reports
- **THEN** the UI MUST request the next `offset` from the API and MUST NOT load all reports into memory

#### Scenario: Show total count

- **WHEN** build reports list loads
- **THEN** the UI MUST display total matching report count from the API `total` field

### Requirement: Build reports date range filter UI

The Build reports page SHALL provide date range inputs for filtering by report time.

#### Scenario: Apply date filter

- **WHEN** enabling team sets From and To dates and applies filters
- **THEN** the UI MUST pass `reportedAfter` and `reportedBefore` to the list API and reset to page 1

#### Scenario: Clear date filter

- **WHEN** enabling team clears date inputs
- **THEN** the UI MUST omit date parameters and show unbounded results (subject to other filters)

### Requirement: Build reports pagination in URL

Pagination state on Build reports SHALL be reflected in URL search parameters for shareable views.

#### Scenario: Bookmark page

- **WHEN** user opens a URL with `page` and `pageSize` query params on Build reports
- **THEN** the UI MUST restore the same page of results
