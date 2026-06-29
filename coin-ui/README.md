# coin-ui — Control Plane admin SPA

React + Vite + TypeScript + Tailwind v4.

## Local dev

```bash
cd coin-ui
npm install
npm run dev   # http://localhost:5173 → proxy /api → coin-api:8090
```

Login: API key / SSO (`VITE_OIDC_*`) / «Пропустить» при `AUTH_DISABLED=true`.

## Docker

```bash
cd docker
make coin-ui-up   # http://localhost:8091
```

## Pages & nav (sidebar)

Навигация — **левый sidebar** с группами. Аудитория: enabling/platform team.

| Group | Route | Содержание | RBAC |
|-------|-------|------------|------|
| **Overview** | `/` | Dashboard (status, versions, stats) | reader+ |
| **Fleet** | `/projects` | Projects registry + canary mode | reader+ |
| | `/build-reports` | Build reports list | reader+ |
| **Golden Paths** | `/gp` | GP Profiles catalog | reader+ |
| | `/gp/:name` | GP hub (Overview, Releases, Policy, Canary) | reader+ |
| | `/gp/:name/releases/:version` | Release detail, composition, blast radius | reader+ |
| | `/gp/new` | New GP profile (name + description) | publisher+ |
| | `/gp/:name/releases/new-draft` | New draft snapshot (gp-content + branching-model) | publisher+ |
| | `/resolve` | Resolve preview + canary debug | reader+ |
| **Platform** | `/platform/runtime` | Agent stack profiles catalog | reader+ |
| | `/platform/runtime/:name` | Agent hub (Overview, Releases) | reader+ |
| | `/platform/runtime/new` | New agent profile | publisher+ |
| | `/platform/runtime/:name/releases/:version` | Agent release detail + derived executor | reader+ |
| | `/platform/build-stacks` | Build stack profiles catalog | reader+ |
| | `/platform/build-stacks/:name` | Build stack hub | reader+ |
| | `/platform/build-stacks/new` | New build stack profile | publisher+ |
| | `/platform/branching-models` | Branching model profiles catalog | reader+ |
| | `/platform/branching-models/:name` | Branching model hub | reader+ |
| | `/platform/branching-models/new` | New branching model profile | publisher+ |
| | `/platform/build-stacks/:name/:version/edit` | gp-content editor (validate → publish) | publisher+ |
| | `/platform/branching-models/:name/:version/edit` | branching-model editor | publisher+ |
| | `/platform/runtime/:name/:version/edit` | agent metadata catch-up | publisher+ |
| | `/components/:type/:name` | redirect → family hub | reader+ |
| **Admin** | `/audit` | Audit log | admin |

**Redirects:** `/branching-models` → `/platform/branching-models`, `/components` → `/platform/runtime`, `/platform/components` → `/platform/runtime`, `/releases` → `/gp`, `/catalog` → `/gp`, `/canary` → `/gp`, `/releases/:n/:v` → `/gp/:n/releases/:v`, `/releases/new-gp` → `/gp/new`, `/releases/publish` → `/gp/:name/releases/new-draft` (с `?name=`), `/platform-settings` → `/audit`.

Header: **API docs ↗** → `/api/docs/` (Swagger UI через proxy).

User guide: [docs/coin-ui-user-guide.md](../docs/coin-ui-user-guide.md)

## OpenAPI types

```bash
make openapi-ui
# или: cd docker && make openapi-ui
```

## Env

| Var | Default | Description |
|-----|---------|-------------|
| `VITE_API_BASE` | `/api` | API prefix (nginx proxy в Docker) |
| `VITE_OIDC_AUTHORITY` | — | OIDC issuer (corp SSO) |
| `VITE_OIDC_CLIENT_ID` | — | SPA client ID |
