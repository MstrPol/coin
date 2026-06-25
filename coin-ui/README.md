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
| **Platform** | `/platform/runtime` | agent/executor catalog | reader+ |
| | `/platform/build-stacks` | gp-content catalog | reader+ |
| | `/platform/branching-models` | Branching models catalog | reader+ |
| | `/platform/components` | Legacy aggregate (deprecated) | reader+ |
| | `/components/:type/:name` | Component detail + publish | reader+ |
| | `/studio` | Component Studio | publisher+ (sidebar footer shortcut) |
| | `/studio/:type/:name/:version` | Editor + validate → canary | publisher+ |
| **Admin** | `/platform-settings` | Nexus settings | admin (edit publisher+) |
| | `/audit` | Audit log | admin |

**Redirects:** `/branching-models` → `/platform/branching-models`, `/components` → `/platform/components`, `/releases` → `/gp`, `/catalog` → `/gp`, `/canary` → `/gp`, `/releases/:n/:v` → `/gp/:n/releases/:v`, `/releases/new-gp` → `/gp/new`, `/releases/publish` → `/gp/:name/releases/new-draft` (с `?name=`).

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
