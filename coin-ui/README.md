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

## Pages

| Route | Содержание | RBAC |
|-------|------------|------|
| `/` | Dashboard | reader+ |
| `/projects` | Projects + filters | reader+ |
| `/releases` | GP releases list | reader+ |
| `/releases/:name/:version` | Detail + blast radius | reader+ |
| `/releases/publish` | Publish wizard | publisher+ |
| `/components` | Component registry | reader+ |
| `/audit` | Audit log | reader+ |

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
