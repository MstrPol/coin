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
    "destinations": {
      "imageRegistryPrefix": "localhost:8082/coin-docker",
      "buildCacheEnabled": true,
      "artifactRepositoryBase": "http://nexus:8081/repository/maven-releases"
    },
    "composition": {
      "agent": "1.0.0",
      "gp-content": "1.0.0",
      "branching-model": "1.0.0"
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
| `publisher` | reader + POST publish GP/components |
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

Версии компонентов (agent, gp-content, branching-model, …):

| Endpoint | Назначение |
|----------|------------|
| `POST .../versions/drafts` | Создать draft (Platform editor) |
| `PATCH .../versions/{v}` | Редактировать draft (metadata, contentRef) |
| `POST .../versions/{v}/validate-package` | Server-side validation draft package |
| `POST .../versions/{v}/register-package` | Draft bodies (PG) → content_ref v2 (без Nexus `package.url`) |
| `POST .../versions/{v}/promote` | draft → published (Nexus upload + immutable package) |
| `PUT .../versions/{v}/artifacts/*` | Draft artifact bodies (PG, Q1) |
| `POST .../versions` | Прямой publish → published (legacy CI scripts) |

**content_ref v2** (Component Package Model): JSON Schema в `schemas/content-ref.v2.schema.json` и `schemas/package.manifest.schema.json`. При `PATCH`/создании draft v2-конверт валидируется в store; legacy `artifactKey` refs допускаются до миграции.

**Resolve visibility** (`ComponentResolveMode`):

| Mode | Когда | Component statuses |
|------|-------|-------------------|
| `stable` | product CI, stable channel | `published` |
| `canary` | canary channel / pilot | `published`, `canary` |
| `admin` | GP draft snapshot preview | `published`, `canary`, `draft` |

CI repos (`coin-executor`, `coin-gp-content`, `coin-lib`) отчитывают версии в API после publish артефакта.
Runtime agent image: `agent/coin-agent` через `coin-executor/scripts/publish-agent.sh`.

Nexus (manifest cache, component packages): env `NEXUS_URL`, `NEXUS_MAVEN_RELEASES`, `NEXUS_MAVEN_SNAPSHOTS` — не operator UI.

## Layout

```
cmd/coin-api/          HTTP server
internal/manifest/     Manifest builder
internal/store/        PostgreSQL access
internal/resolve/      Resolve service
migrations/            goose SQL
openapi/v1.yaml        API contract
schemas/               content_ref v2 + package.manifest JSON Schema
manifest.schema.json   Resolved manifest JSON Schema
```

## API docs

Swagger UI: http://localhost:8090/docs/

OpenAPI: `GET /openapi/v1.yaml`
