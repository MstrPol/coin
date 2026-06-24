## ADDED Requirements

### Requirement: Projects CSV export endpoint

The coin-api SHALL provide a CSV export for projects matching list filters.

#### Scenario: Export all filtered projects

- **WHEN** client calls `GET /v1/admin/projects/export` with the same filters as the list endpoint
- **THEN** the API MUST respond with `Content-Type: text/csv` and all matching project rows (not limited to one page)

#### Scenario: CSV columns

- **WHEN** export completes
- **THEN** the CSV MUST include at minimum: name, groupId, artifactId, gitRepoName, gitRepoUrl, goldenPath, version, canaryMode, branch, lastBuildAt

### Requirement: Build reports CSV export endpoint

The coin-api SHALL provide a CSV export for build reports matching list filters including date range.

#### Scenario: Export all filtered reports

- **WHEN** client calls `GET /v1/admin/build-reports/export` with project, goldenPath, result, reportedAfter, and reportedBefore filters
- **THEN** the API MUST respond with `Content-Type: text/csv` and all matching report rows

#### Scenario: CSV columns for reports

- **WHEN** export completes
- **THEN** the CSV MUST include at minimum: id, project, goldenPath, version, resolvedVersion, result, channel, branch, buildUrl, failedStage, reportedAt

### Requirement: UI CSV download action

Fleet table pages SHALL offer an Export CSV control that downloads the full filtered dataset.

#### Scenario: Export from Projects page

- **WHEN** enabling team clicks Export CSV on Projects with active filters
- **THEN** the UI MUST download a CSV file containing all projects matching those filters via the export endpoint

#### Scenario: Export from Build reports page

- **WHEN** enabling team clicks Export CSV on Build reports with active filters including date range
- **THEN** the UI MUST download a CSV file containing all matching build reports via the export endpoint

#### Scenario: Export does not require loading all pages in browser

- **WHEN** export is triggered
- **THEN** the UI MUST NOT paginate through list endpoints client-side to assemble the CSV
