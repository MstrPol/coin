# coin-ui — руководство пользователя

**URL (local):** http://localhost:8091

## Назначение

Operator UI для Control Plane: fleet analytics, GP releases, политика версий, canary, resolve preview, component registry, audit.

## Вход

1. Открыть http://localhost:8091
2. **API key** — `COIN_ADMIN_API_KEY` / `COIN_PUBLISHER_API_KEY` / `COIN_READER_API_KEY`
3. **SSO** — если заданы `VITE_OIDC_AUTHORITY` + `VITE_OIDC_CLIENT_ID` (corp)
4. **Пропустить** — local dev при `AUTH_DISABLED=true` (роль admin)

После входа header показывает subject, роль и ссылку **API docs ↗** (Swagger UI).

### RBAC (local demo)

| Key (docker/.env) | Роль | Publish GP / components |
|-------------------|------|-------------------------|
| `dev-local-admin-key` | admin | ✅ |
| `dev-local-publisher-key` | publisher | ✅ |
| `dev-local-reader-key` | reader | ❌ (403 на POST/PUT/PATCH) |

Key или OIDC access token хранится в `localStorage`.

## Nav (sidebar)

Группы в левом sidebar:

| Group | Пункт | Route | RBAC |
|-------|-------|-------|------|
| Overview | Dashboard | `/` | reader+ |
| Fleet | Projects | `/projects` | reader+ |
| Fleet | Build reports | `/build-reports` | reader+ |
| Golden Paths | GP Profiles | `/gp` | reader+ |
| Golden Paths | Resolve | `/resolve` | reader+ |
| Platform | Runtime | `/platform/runtime` | reader+ |
| Platform | Build stacks | `/platform/build-stacks` | reader+ |
| Platform | Branching models | `/platform/branching-models` | reader+ |
| Admin | Audit | `/audit` | admin |
| Footer | — | Studio удалён; `/studio` → redirect на Platform |

**Redirects:** `/branching-models` → `/platform/branching-models`, `/components` → `/platform/components`, `/releases` → `/gp`, `/catalog` → `/gp`, `/canary` → `/gp`, `/releases/:n/:v` → `/gp/:n/releases/:v`, `/releases/new-gp` → `/gp/new`, `/releases/publish` → `/gp/:name/releases/new-draft` (с `?name=`), `/platform-settings` → `/audit`.

**Publish flows** — внутри GP hub (кнопки на hub / Releases tab), не в sidebar.

## Страницы

### Dashboard

- Статус coin-api (`GET /ready`) + semver **coin-api** и **coin-ui**
- Счётчики: projects, **stale projects**, GP releases, build reports, golden paths
- Клик по карточке → соответствующая страница

### Projects

Регистрация при **первом build report**, обновление при каждом `POST /v1/builds/report`.

Колонки: name, groupId, artifactId, git repo (ссылка), GP pin, version pin, canary mode, last build, branch.

Фильтры: `goldenPath`, `version`, `stale` (`/projects?stale=1` — без билда >90 дней).

**Пагинация:** server-side (`limit`/`offset`, default 50). URL: `page`, `pageSize`. Счётчик «N из total».

**Export CSV** — полный набор по текущим фильтрам (`GET /v1/admin/projects/export`).

**Canary mode** per project: `default` | `canary` | `stable` — override для pin `*` (см. [canary.md](canary.md)).

### Build reports

История `POST /v1/builds/report`: project, GP, pin, resolved version, result, channel, branch, build URL, время.

Фильтры: project, goldenPath, result, **даты** (`reportedAfter`, `reportedBefore` — `YYYY-MM-DD`).

**Пагинация:** server-side, URL `page` + `pageSize`.

**Export CSV** — все matching reports по фильтрам (`GET /v1/admin/build-reports/export`).

### GP Profiles (`/gp`)

Каталог Golden Path profiles (не flat-список всех версий). Колонки: description, latest stable/canary, release count.

- **Open** → GP hub (`/gp/:name`) с вкладками Overview, Releases, Policy, Canary
- **New profile** (`/gp/new`) — `name` + optional `description`; composition — в первом draft

### GP hub (`/gp/:name`)

Entity-centric view одного GP:

| Вкладка | Содержание |
|---------|------------|
| Overview | Description, policy summary, quick links |
| Releases | Список releases + drafts для этого GP |
| Policy | latest / minimum / deprecated, resolve preview |
| Canary | Rollout %, health, resolve preview |

**Publisher actions:** New draft (`/gp/:name/releases/new-draft`) — 3 pickers: agent stack, gp-content, branching-model. Stable release — только promote draft на release detail.

**Draft lifecycle:** пока `status=draft` — composition редактируется на release detail (**Save composition**), draft можно удалить. После promote — read-only, без Save/Delete.

### GP release detail (`/gp/:name/releases/:version`)

Composition (agent, gp-content, branching-model). **Draft:** editable form + **Save composition**, Promote (блокируется при draft pins), Delete draft. **Published:** read-only. Ссылки на Platform editor для gp-content / branching-model.

### Resolve preview

Тест resolve engine: pin + project из registry → resolved version, channel, manifest JSON.

При выбранном project — панель **Canary status** (audience, mode, bucket, rollout).

Override **auto | stable | canary** — только для preview (`forceChannel`), не меняет project в БД.

### Components (Platform)

Каталоги по ролям в composition — **Platform** в sidebar. Каждое семейство: **профили** → **hub** (Overview + Releases), по образцу GP.

- **Runtime** (`/platform/runtime`) — agent stack profiles; hub `/platform/runtime/:name`
- **Build stacks** (`/platform/build-stacks`) — gp-content profiles; hub `/platform/build-stacks/:name`
- **Branching models** (`/platform/branching-models`) — branching-model profiles; hub `/platform/branching-models/:name`

Primary actions: **New profile** на каталоге, **New draft** на hub.

Legacy: `/platform/components`, `/components/agent/:name` → hub. `/platform/jenkins-lib` → `/platform/runtime`.

**Editors:** gp-content / branching-model — `/platform/.../:name/:version/edit` (validate → register → promote). Agent — metadata catch-up `/platform/runtime/:name/:version/edit`, CI path через `publish-agent.sh` (draft register + promote).

**Release detail:** `/platform/{family}/:name/releases/:version` — для agent показывается derived `executor/coin-executor@version`.

### Audit log

Журнал mutations: `publish_gp_release`, `publish_component_version`, `update_platform_settings`, …

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
| Swagger UI | `/api/docs/` → coin-api `/docs/` |

## Ограничения (pilot)

- Corp rollout — [prod-repo-split.md](runbooks/prod-repo-split.md) после corp gate
- Auto-rollback canary — не реализован (только health signal)
- Manifest tree viewer — вне scope (resolve preview — JSON)

## Связанные документы

- [canary.md](canary.md)
- [fleet-analytics-pm.md](how-to/fleet-analytics-pm.md)
- [openapi.md](openapi.md)
- [coin-ui/README.md](../coin-ui/README.md)
- [local-dev-control-plane.md](how-to/local-dev-control-plane.md)
