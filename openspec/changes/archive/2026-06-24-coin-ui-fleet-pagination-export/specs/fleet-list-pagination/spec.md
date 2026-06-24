## ADDED Requirements

### Requirement: Paginated list response

List endpoints for fleet tables SHALL return server-side pagination metadata alongside items.

#### Scenario: Projects list with total

- **WHEN** client calls `GET /v1/admin/projects` with `limit` and `offset`
- **THEN** the response MUST include `items`, `total`, `limit`, and `offset`

#### Scenario: Build reports list with total

- **WHEN** client calls `GET /v1/admin/build-reports` with `limit` and `offset`
- **THEN** the response MUST include `items`, `total`, `limit`, and `offset`

### Requirement: Projects query pagination

The projects list endpoint SHALL support `limit` and `offset` query parameters with server-side slicing.

#### Scenario: Default page size

- **WHEN** client omits `limit` on projects list
- **THEN** the API MUST default to `limit=50` and MUST NOT return the full fleet in one response

#### Scenario: Maximum page size

- **WHEN** client requests `limit` greater than 500
- **THEN** the API MUST cap the page size at 500

### Requirement: Build reports date range filter

The build reports list endpoint SHALL accept optional date range filters on `reportedAt`.

#### Scenario: Filter by start date

- **WHEN** client passes `reportedAfter=2026-01-01`
- **THEN** the API MUST return only reports with `reportedAt` on or after that date (start of day UTC)

#### Scenario: Filter by end date

- **WHEN** client passes `reportedBefore=2026-01-31`
- **THEN** the API MUST return only reports with `reportedAt` on or before that date (end of day UTC)

#### Scenario: Combined date range

- **WHEN** both `reportedAfter` and `reportedBefore` are provided
- **THEN** the API MUST return only reports within the inclusive range
