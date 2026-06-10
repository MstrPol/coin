# coin-ui — руководство пользователя

**Ticket:** P2-08, P3-03  
**URL (local):** http://localhost:8091

## Назначение

Operator UI для Control Plane: fleet analytics, GP publish/draft, catalog pins, canary и resolve preview.

## Вход

1. Открыть http://localhost:8091
2. **API key** — `COIN_ADMIN_API_KEY` / `COIN_PUBLISHER_API_KEY` / `COIN_READER_API_KEY`
3. **SSO** — если заданы `VITE_OIDC_AUTHORITY` + `VITE_OIDC_CLIENT_ID` (corp)
4. **Пропустить** — local dev при `AUTH_DISABLED=true` (роль admin)

После входа header показывает subject и роль. **Publish** виден только `publisher`/`admin`.

### RBAC (local demo)

| Key (docker/.env) | Роль | Publish GP |
|-------------------|------|------------|
| `dev-local-admin-key` | admin | ✅ + fleet scan |
| `dev-local-publisher-key` | publisher | ✅ |
| `dev-local-reader-key` | reader | ❌ (403) |

Key или OIDC access token хранится в `localStorage`.

## Страницы

### Dashboard

- Статус coin-api (`/ready`)
- Счётчики: projects, GP releases, build reports, golden paths
- Клик по карточке → Projects / GP Releases

### Projects

Таблица projects с **последним GP binding** (из build report или scanner).

Фильтры: `goldenPath`, `version`. URL: `/projects?goldenPath=go-app&version=1.0.0`

**Canary mode** per project: `default` | `canary` | `stable` — override для pin `*` (см. [canary.md](canary.md)).

### GP Releases

Список published + draft releases из `gp_releases`.

- **Publish** — wizard ([`/releases/publish`](http://localhost:8091/releases/publish)): create draft snapshot, direct publish, promote draft
- **Detail** — composition, artifact editor (draft only), blast radius (published)

### Catalog

Pointer status для GP: `*`, `=latest`, `~`, `^`, `canary:latest`.

Редактирование `latest`, `latest_canary`, `minimum`, deprecated list (publisher/admin).

### Resolve preview

Тест resolve engine: pin + optional `project` → resolved version, channel header, manifest hash preview.

### Canary

Slider `canary_percent`, preview audience (сколько projects в canary bucket), health badge по build reports.

Пороги degraded/critical — из `canary_policy`.

### Audit log

Журнал append-only mutations: `publish_gp_release`, `publish_component_version`.

Фильтры по `entityType` / `action`, pagination, раскрытие JSON payload.

### Components

Component registry (read-only): type, name, latest version, count versions.

## Запуск

```bash
cd docker
make coin-ui-up    # coin-api + coin-ui
```

Dev без Docker:

```bash
cd coin-ui && npm run dev   # :5173, proxy /api → :8090
```

## API proxy

| Окружение | API base |
|-----------|----------|
| Docker | `/api` → nginx → `coin-api:8090` |
| Vite dev | `/api` → vite proxy → localhost:8090 |

## Ограничения (pilot)

- Component publish — curl/Admin API (wizard — GP only)
- Fleet scanner — [scanner-ops.md](runbooks/scanner-ops.md) (`make scan-fleet`, CronJob)
- Auto-rollback canary — не реализован (только health signal)
- Corp rollout — [prod-repo-split.md](runbooks/prod-repo-split.md) после corp gate

## Связанные документы

- [canary.md](canary.md)
- [fleet-analytics-pm.md](how-to/fleet-analytics-pm.md)
- [coin-ui/README.md](../coin-ui/README.md)
- [local-dev-control-plane.md](how-to/local-dev-control-plane.md)
