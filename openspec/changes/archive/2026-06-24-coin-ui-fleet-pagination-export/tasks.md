## 1. coin-api — pagination

- [x] 1.1 `ListProjects`: `limit`, `offset`, `CountProjects` с теми же фильтрами
- [x] 1.2 `listProjects` handler: response `{ items, total, limit, offset }`
- [x] 1.3 `ListBuildReports`: `reportedAfter`, `reportedBefore`, `CountBuildReports`
- [x] 1.4 `listBuildReports` handler: response `{ items, total, limit, offset }`
- [x] 1.5 OpenAPI: query params + response schema для обоих list endpoints

## 2. coin-api — CSV export

- [x] 2.1 `GET /v1/admin/projects/export` — streaming CSV, те же фильтры
- [x] 2.2 `GET /v1/admin/build-reports/export` — streaming CSV + date filters
- [x] 2.3 OpenAPI + store tests для count/date filters

## 3. coin-ui — shared

- [x] 3.1 `PaginatedListResponse<T>` type + `api.projects` / `api.buildReports` с limit/offset/total
- [x] 3.2 `TablePagination` component (page, pageSize, total, URL sync)
- [x] 3.3 `downloadCsv(url, filename)` helper с auth header

## 4. coin-ui — Projects

- [x] 4.1 Server-side pagination + total count display
- [x] 4.2 Export CSV button (current filters)
- [x] 4.3 URL params: `page`, `pageSize` (+ existing goldenPath/version/stale)

## 5. coin-ui — Build reports

- [x] 5.1 Date range inputs (from/to) + API params
- [x] 5.2 Server-side pagination + total count
- [x] 5.3 Export CSV button (filters + dates)
- [x] 5.4 URL params: `page`, `pageSize`, `reportedAfter`, `reportedBefore`

## 6. Docs & acceptance

- [x] 6.1 Update `docs/coin-ui-user-guide.md` (pagination, dates, CSV)
- [x] 6.2 Manual smoke: pagination + CSV + date filter (pilot API curl)
- [x] 6.3 `openspec validate coin-ui-fleet-pagination-export --strict`
