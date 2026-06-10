# coin-api

Coin Control Plane HTTP API: Resolve manifest, build report, admin (phases 2+).

## Local dev

```bash
# from docker/
make coin-api-up
curl http://localhost:8090/health
curl http://localhost:8090/v1/golden-paths/go-app/versions/1.0.0/manifest

# Nexus fallback (after resolve warmed cache): pointer → blob
BASE=http://localhost:8081/repository/coin-manifests
curl -fsS "${BASE}/pointers/go-app/%3D1.0.0.json" | jq .

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
| `GITEA_URL` | `http://gitea:3000` | Fleet scanner (list repos) |
| `GITEA_ORG` | `coin` | Gitea user/org for fleet scanner |
| `GIT_EXPORT_DISABLED` | `true` | Legacy Gitea tag export removed (PF-17) |
| `NEXUS_URL` | `http://nexus:8081` | Manifest cache upload |
| `NEXUS_MANIFEST_REPO` | `coin-manifests` | Raw repo name |

Migrations run automatically on startup (goose).

## RBAC (P4-02)

| Role | Admin API |
|------|-----------|
| `reader` | GET `/v1/admin/*` |
| `publisher` | reader + POST publish GP/components |
| `admin` | publisher + POST `/v1/admin/scan` |

```bash
curl -H "X-API-Key: dev-local-reader-key" http://localhost:8090/v1/admin/me
# reader → 403 on POST publish

curl -H "X-API-Key: dev-local-publisher-key" -X POST .../golden-paths/go-app/versions ...
```

OIDC: `OIDC_ENABLED=true` + Bearer JWT with `roles` claim. See [coin-ui user guide](../docs/coin-ui-user-guide.md).

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

## Fleet scanner (P3-01)

```bash
# from host (Gitea on localhost:3000)
export DATABASE_URL="postgres://coin:coin@localhost:5432/coin?sslmode=disable"
export GITEA_URL=http://localhost:3000
export GITEA_USER=coin GITEA_PASSWORD=coin GITEA_ORG=coin

go run ./cmd/scanner
go run ./cmd/scanner -force   # ignore incremental SHA cache
```

Docker: `cd docker && make scan-fleet`

## Fleet scanner API + CronJob (P3-02)

```bash
# Trigger scan via API (updates Prometheus metrics on coin-api /metrics)
curl -X POST http://localhost:8090/v1/admin/scan \
  -H "X-API-Key: dev-local-admin-key"

curl http://localhost:8090/metrics | grep coin_scan
```

K8s (local k3s in docker-compose):

```bash
cd docker
make endpoints              # coin-api Endpoints in k3s
make scan-cronjob-apply     # nightly CronJob 02:00
make scan-cronjob-run       # one-off Job + logs
```

Metrics: `coin_scan_duration_seconds`, `coin_repos_scanned`, `coin_scan_repos_total`,
`coin_scan_repos_skipped`, `coin_scan_repos_failed`, `coin_scan_last_success_timestamp`.
