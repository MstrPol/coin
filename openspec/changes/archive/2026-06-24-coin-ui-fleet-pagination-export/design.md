## Context

**Текущее состояние:**

| Endpoint | Пагинация | Total | Date filter | UI |
|----------|-----------|-------|-------------|-----|
| `GET /v1/admin/projects` | ❌ все rows | ❌ | — | все rows в DOM |
| `GET /v1/admin/build-reports` | `limit`+`offset` в store, UI шлёт только `limit=100` | ❌ | ❌ | без pager |

List response: `{ items: T[] }` — без `total`.

**Ограничения:** local pilot; fleet analytics для enabling team; PostgreSQL SoT.

## Goals / Non-Goals

**Goals:**

- Server-side pagination: UI запрашивает одну страницу за раз.
- `total` для отображения «N записей», номеров страниц.
- CSV export полного набора по **текущим фильтрам** без удержания всего в памяти браузера.
- Build reports: фильтр по диапазону `reported_at`.

**Non-goals:**

- Cursor-based pagination (offset достаточен для pilot scale).
- Server-side column sorting.
- Изменение auth model для export.

## Decisions

### D1: List response shape (additive)

```json
{ "items": [...], "total": 1234, "limit": 50, "offset": 0 }
```

Обратная совместимость: старые клиенты читают только `items`. OpenAPI обновить.

### D2: Projects pagination в PostgreSQL

Добавить `LIMIT`/`OFFSET` к существующему `ListProjects` + отдельный `COUNT(*)` с теми же WHERE (goldenPath, version, stale).

Default `limit=50`, max `500` (как build-reports).

### D3: Build reports date filters

Query params (RFC3339 date or `YYYY-MM-DD`):

- `reportedAfter` — inclusive start (00:00:00 UTC)
- `reportedBefore` — inclusive end (23:59:59 UTC)

SQL: `br.reported_at >= $after AND br.reported_at <= $before`.

### D4: CSV export — dedicated streaming endpoints

| Endpoint | Filters |
|----------|---------|
| `GET /v1/admin/projects/export` | те же query: goldenPath, version, stale |
| `GET /v1/admin/build-reports/export` | project, goldenPath, result, reportedAfter, reportedBefore |

Response: `Content-Type: text/csv`, `Content-Disposition: attachment; filename="projects-YYYYMMDD.csv"`.

Реализация: `pgx` row iterator → `csv.Writer` → `http.ResponseWriter` (chunked). Без materialize всего slice в RAM.

Колонки CSV = все поля JSON row (заголовки snake_case или camelCase как в API — **camelCase** для consistency с JSON).

### D5: UI pagination component

Shared `TablePagination` в coin-ui:

- page size select: 25 / 50 / 100 (default 50)
- prev / next + «Страница X из Y»
- sync `page` + `pageSize` в URL searchParams (bookmarkable)

При смене фильтра — сброс на page 1.

### D6: Export CSV button

- Вызывает export endpoint с **текущими фильтрами** (не текущей страницей).
- `fetch` + blob download или `window.open` с API key header — предпочтительно `fetch` + `URL.createObjectURL` (SPA уже хранит key в localStorage).
- Loading state + error toast/message.

### D7: Build reports date UI

Два `<input type="date">` (from / to) + Apply вместе с остальными фильтрами. Пустые = без ограничения.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| OFFSET slow на очень больших offset | Приемлемо для pilot; cursor — follow-up |
| COUNT(*) на projects с LATERAL join | COUNT на projects без join; last build — только в list query |
| CSV timeout на nginx | Увеличить proxy read timeout для export в docker nginx при необходимости |
| Export без rate limit | reader+ RBAC; pilot only |

## Migration Plan

1. Deploy coin-api (additive query params).
2. Deploy coin-ui (использует новые поля).
3. Rollback: UI fallback на items-only если total отсутствует (не нужен при joint deploy).

## Open Questions

| # | Вопрос | Статус | Lean |
|---|--------|--------|------|
| Q1 | Default page size 50 vs 25? | ⏳ | 50 |
| Q1 | Имя export endpoint `/export` vs `?format=csv` | ⏳ | `/export` sub-resource |
