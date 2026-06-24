# coin-ui — руководство пользователя

**URL (local):** http://localhost:8091

## Назначение

Operator UI для Control Plane: fleet analytics, GP releases, политика версий, canary, resolve preview, component registry, platform settings.

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
| Golden Paths | Releases | `/releases` | reader+ |
| Golden Paths | GP Policy | `/catalog` | reader+ |
| Golden Paths | Canary | `/canary` | reader+ |
| Golden Paths | Resolve | `/resolve` | reader+ |
| Platform | Runtime | `/platform/runtime` | reader+ |
| Platform | Build stacks | `/platform/build-stacks` | reader+ |
| Platform | Branching models | `/platform/branching-models` | reader+ |
| Platform | Jenkins library | `/platform/jenkins-lib` | reader+ |
| Admin | Platform settings | `/platform-settings` | admin |
| Admin | Audit | `/audit` | admin |
| Footer | Studio | `/studio` | publisher+ |

**Redirects:** `/branching-models` → `/platform/branching-models`, `/components` → `/platform/components` (legacy aggregate).

**Publish wizard** — только с GP Releases (кнопка Publish), не в sidebar. Route: `/releases/publish`.

## Страницы

### Dashboard

- Статус coin-api (`GET /ready`) + semver **coin-api** и **coin-ui**
- Счётчики: projects, **stale projects**, GP releases, build reports, golden paths
- Клик по карточке → соответствующая страница

### Projects

Регистрация при **первом build report**, обновление при каждом `POST /v1/builds/report`.

Колонки: name, groupId, artifactId, git repo (ссылка), GP pin, version pin, canary mode, last build, branch.

Фильтры: `goldenPath`, `version`, `stale` (`/projects?stale=1` — без билда >90 дней).

**Canary mode** per project: `default` | `canary` | `stable` — override для pin `*` (см. [canary.md](canary.md)).

### Build reports

История `POST /v1/builds/report`: project, GP, pin, resolved version, result, channel, branch, build URL, время.

Фильтры: project, goldenPath, result.

### GP Releases

Список published + draft releases. Dropdown-фильтр по GP.

- **Publish** — wizard (`/releases/publish`): draft snapshot, direct publish, promote
- **Detail** — composition, artifact editor (draft), blast radius (published)

### GP Policy (бывш. Catalog)

Политика версий GP: latest stable, latest canary, minimum, deprecated.

Превью resolve для pin `*` — stable и canary линии (★).

### Resolve preview

Тест resolve engine: pin + project из registry → resolved version, channel, manifest JSON.

При выбранном project — панель **Canary status** (audience, mode, bucket, rollout).

Override **auto | stable | canary** — только для preview (`forceChannel`), не меняет project в БД.

### Canary

Slider `canary_percent`, preview audience, health badge по build reports.

### Components (Platform)

Каталоги по ролям в composition — **Platform** в sidebar:

- **Runtime** (`/platform/runtime`) — `agent`, `executor`
- **Build stacks** (`/platform/build-stacks`) — `gp-content`
- **Branching models** (`/platform/branching-models`) — lifecycle draft → canary → published
- **Jenkins library** (`/platform/jenkins-lib`) — `lib` / coin-lib

Legacy aggregate: `/platform/components` (redirect с `/components`).

Detail: `/components/:type/:name` — версии, metadata/contentRef, GP usage, publish (publisher).

**GP release detail** — вкладка **Build stack**: gp-content версии для profile name, ссылки в Studio.

**Component Studio** — `/studio` (publisher, shortcut в footer sidebar). Authoring: validate → register (PG) → canary → promote (Nexus).

### Platform settings

Глобальные настройки Nexus (`nexus.mavenBase`, `nexus.credentialsId`) — бывший `platform.yaml`.

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
