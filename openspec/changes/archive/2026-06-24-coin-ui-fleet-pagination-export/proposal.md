## Why

Страницы **Projects** и **Build reports** в coin-ui загружают все записи разом (projects — без limit; build reports — `limit=100` без offset/total). При росте fleet (сотни–тысячи repos и десятки тысяч reports) UI тормозит, а enabling team не может выгрузить полный срез для анализа. Нужна server-side пагинация и CSV export без загрузки всего датасета в браузер.

## What Changes

- **coin-api:** `GET /v1/admin/projects` — `limit`, `offset`, `total` в ответе (сейчас отдаёт весь список).
- **coin-api:** `GET /v1/admin/build-reports` — `total`, фильтры `reportedAfter` / `reportedBefore` (даты); `offset` уже в store, задокументировать в OpenAPI.
- **coin-api:** CSV export endpoints с теми же фильтрами, streaming response (`text/csv`).
- **coin-ui:** пагинация на Projects и Build reports (page size, prev/next, total count).
- **coin-ui:** кнопка **Export CSV** на обеих страницах — скачивание полного набора по текущим фильтрам через API export.
- **Build reports UI:** date range picker (from/to) в дополнение к project/GP/result.

## Capabilities

### New Capabilities

- `fleet-list-pagination`: server-side pagination contract (projects + build-reports), `total` в list response.
- `fleet-csv-export`: CSV export endpoints и UI download по активным фильтрам.
- `build-reports-fleet`: date-range фильтрация и пагинация на странице Build reports.

### Modified Capabilities

- `fleet-project-hub`: пагинация и CSV export на странице Projects.

## Impact

- **coin-api:** `store/registry.go`, `store/build_reports.go`, handlers, OpenAPI.
- **coin-ui:** `Projects.tsx`, `BuildReports.tsx`, shared `Pagination` + `api.ts`, types.
- **docs:** `coin-ui-user-guide.md` — фильтры и export.
- **Non-goals:** client-side-only pagination; infinite scroll; column sort; corp fleet rollout.

## Non-goals

- Загрузка всех строк в SPA с последующей client-side пагинацией.
- Excel/XLSX export.
- Real-time live table updates.
- Изменение схемы PostgreSQL.
- Wave rollout 50/500/1500 repos (corp gate).
