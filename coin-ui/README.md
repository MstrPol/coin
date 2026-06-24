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

## Pages & nav

| Route | Содержание | Nav | RBAC |
|-------|------------|-----|------|
| `/` | Dashboard (status, versions, stats) | Dashboard | reader+ |
| `/projects` | Projects registry + canary mode | Projects | reader+ |
| `/build-reports` | Build reports list | — (Dashboard link) | reader+ |
| `/releases` | GP releases + GP filter | GP Releases | reader+ |
| `/releases/:name/:version` | Detail, artifacts, blast radius | — | reader+ |
| `/releases/publish` | Publish wizard | — (кнопка на GP Releases) | publisher+ |
| `/catalog` | GP Policy (version policy) | GP Policy | reader+ |
| `/promote` | Promote canary → stable wizard | — | publisher+ |
| `/resolve` | Resolve preview + canary debug | Resolve | reader+ |
| `/canary` | Canary policy + health | Canary | reader+ |
| `/components` | Component list | Components | reader+ |
| `/components/:type/:name` | Component detail + publish | — | reader+ |
| `/studio` | Component Studio — новый draft | Studio | publisher+ |
| `/studio/:type/:name/:version` | Editor + validate → canary + pilot promote gate | — | publisher+ |
| `/platform-settings` | Nexus platform settings | Platform | reader+ / edit publisher+ |
| `/audit` | Audit log | Audit | reader+ |

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
