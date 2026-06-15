# coin-api

Coin Control Plane HTTP API: Resolve manifest, build report, admin (phases 2+).

## Local dev

```bash
# from docker/
make coin-api-up
curl http://localhost:8090/health
curl http://localhost:8090/v1/golden-paths/go-app/versions/1.0.0/manifest

# Nexus fallback (after resolve warmed cache): pointer → blob
BASE=http://localhost:8081/repository/maven-snapshots
curl -fsS "${BASE}/coin/manifest/go-app/metadata/go-app-metadata-pin-%3D1.0.0.json" | jq .

# MVP-1 E2E smoke (from docker/)
make e2e-mvp1

# Admin publish (local, AUTH_DISABLED=true — key optional)
curl -X POST http://localhost:8090/v1/admin/components/test/widget/versions \
  -H "Content-Type: application/json" \
  -d '{"version":"1.0.0","metadata":{"note":"pilot"}}'

curl -X POST http://localhost:8090/v1/admin/golden-paths/go-app/versions \
  -H "Content-Type: application/json" \
  -d '{
    "version": "1.0.1",
    "composition": {
      "executor": "0.1.0", "agent": "1.22.5", "pipeline": "2.1.0",
      "validate": "1.0.0", "dockerfile": "1.0.0"
    }
  }'
```

## Env

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | required | PostgreSQL DSN |
| `COIN_API_ADDR` | `:8090` | Listen address |
| `AUTH_DISABLED` | `false` | Local dev: `true` |
| `COIN_API_TOKEN` | — | Bearer token for `/v1/*` (required if auth on) |
| `COIN_ADMIN_API_KEY` | — | Role **admin** (full admin API) |
| `COIN_PUBLISHER_API_KEY` | — | Role **publisher** (read + publish GP/components) |
| `COIN_READER_API_KEY` | — | Role **reader** (GET `/v1/admin/*` only) |
| `OIDC_ENABLED` | `false` | Validate Bearer JWT on admin routes |
| `OIDC_ISSUER_URL` | — | OIDC issuer (Keycloak realm URL) |
| `OIDC_AUDIENCE` | — | JWT audience / client ID |
| `OIDC_ROLES_CLAIM` | `roles` | JWT claim with `[admin,publisher,reader]` |
| `GIT_EXPORT_DISABLED` | `true` | Legacy Gitea tag export removed (PF-17) |
| `NEXUS_URL` | `http://nexus:8081` | Manifest cache upload |
| `NEXUS_MAVEN_RELEASES` | `maven-releases` | Maven2 hosted (releases) |
| `NEXUS_MAVEN_SNAPSHOTS` | `maven-snapshots` | Maven2 hosted (pointers, snapshots) |

Migrations run automatically on startup (goose).

## RBAC (P4-02)

| Role | Admin API |
|------|-----------|
| `reader` | GET `/v1/admin/*` |
| `publisher` | reader + POST publish GP/components + PUT platform settings |
| `admin` | same as publisher |

```bash
curl -H "X-API-Key: dev-local-reader-key" http://localhost:8090/v1/admin/me
# reader → 403 on POST publish

curl -H "X-API-Key: dev-local-publisher-key" -X POST .../golden-paths/go-app/versions ...
```

OIDC: `OIDC_ENABLED=true` + Bearer JWT with `roles` claim. See [coin-ui user guide](../docs/coin-ui-user-guide.md).

## Projects registry

Проекты регистрируются при первом `POST /v1/builds/report` и обновляются при каждом билде.
Fleet scanner удалён (UI-02).

## Component registry SoT

Версии компонентов (agent, executor, …) публикуются через `POST /v1/admin/components/{type}/{name}/versions`.
CI repos (`coin-jenkins-agents`, `coin-executor`) отчитывают версии в API после publish артефакта.

Глобальные настройки Nexus: `GET/PUT /v1/admin/platform/settings`.

## Layout

```
cmd/coin-api/          HTTP server
internal/manifest/     Manifest builder
internal/store/        PostgreSQL access
internal/resolve/      Resolve service
migrations/            goose SQL
openapi/v1.yaml        API contract
manifest.schema.json   Resolved manifest JSON Schema
```

## API docs

Swagger UI: http://localhost:8090/docs/

OpenAPI: `GET /openapi/v1.yaml`
