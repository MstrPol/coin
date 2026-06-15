# Publish GP release (Admin API)

**Ticket:** P2-08  
**Audience:** Platform team  
**Scope:** local stend + подготовка к corp (без wave rollout).

## Prerequisites

- coin-api с Admin API (P2-01…P2-03)
- Компоненты опубликованы в DB (`component_versions`)
- `COIN_ADMIN_API_KEY` или `AUTH_DISABLED=true` (local)

## Шаги

### 1. Проверить composition

Для `go-app` обязательны ключи:

| Key | Component |
|-----|-----------|
| `executor` | coin-executor@{ver} |
| `agent` | go@{ver} |
| `pipeline` | go-build@{ver} |
| `validate` | config@{ver} |
| `dockerfile` | go-runtime@{ver} |

Compatibility matrix проверяется автоматически (pipeline 2.1.x → executor/agent constraints).

### 2. Publish GP

```bash
curl -X POST http://localhost:8090/v1/admin/golden-paths/go-app/versions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${COIN_ADMIN_API_KEY:-dev-local-admin-key}" \
  -d '{
    "version": "1.0.3",
    "composition": {
      "executor": "0.1.0",
      "agent": "1.22.5",
      "pipeline": "2.1.0",
      "validate": "1.0.0",
      "dockerfile": "1.0.0"
    },
    "actor": "platform-team"
  }'
```

**Side-effects:**

1. Row в `gp_releases` + `gp_composition` (append-only)
2. Resolve → canonical manifest → Nexus **blob + pointers** (`latest`, `~`, `^`, `=`)
3. Content artifacts в Nexus (`content/{gp}/{ver}/…`)
4. Audit log

Git export tag в Gitea **не используется** (PF-17) — SoT = PostgreSQL + Nexus.

### 3. Draft / snapshot (optional)

```bash
# Create draft (editable snapshot, PG only — not in fleet pins)
curl -X POST http://localhost:8090/v1/admin/golden-paths/go-app/drafts \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${COIN_ADMIN_API_KEY:-dev-local-publisher-key}" \
  -d '{"version":"1.0.3-snapshot.1","baseVersion":"1.0.2","actor":"platform-team"}'

# Promote draft → published (upload to Nexus, update catalog pointers)
curl -X POST "http://localhost:8090/v1/admin/golden-paths/go-app/versions/1.0.3-snapshot.1/promote" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${COIN_ADMIN_API_KEY:-dev-local-publisher-key}" \
  -d '{"actor":"platform-team"}'
```

coin-ui: [Publish wizard](../coin-ui-user-guide.md) — tabs **Create draft**, **Publish**, **Promote draft**.

### 4. Verify

```bash
# Resolve manifest (url-shaped content refs, orchestration from Nexus)
curl -sf http://localhost:8090/v1/golden-paths/go-app/versions/1.0.3/manifest \
  | jq '{hash: .manifestHash, orch: .orchestration.url}'

# Nexus pointer (exact pin)
SNAPSHOTS=http://localhost:8081/repository/maven-snapshots
curl -sf "${SNAPSHOTS}/coin/manifest/go-app/metadata/go-app-metadata-pin-%3D1.0.3.json" | jq '{manifestHash, blobUrl}'

# Blast radius (coin-ui или curl)
curl -sf -H "X-API-Key: ${COIN_ADMIN_API_KEY:-dev-local-reader-key}" \
  http://localhost:8090/v1/admin/golden-paths/go-app/versions/1.0.3/blast-radius | jq .
```

### 5. Publish component (если нужен новый)

```bash
curl -X POST http://localhost:8090/v1/admin/components/executor/coin-executor/versions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${COIN_ADMIN_API_KEY}" \
  -d '{"version":"0.2.0","metadata":{"url":"http://nexus:8081/.../coin-executor-linux-arm64"}}'
```

Duplicate version → **409**.

## Ошибки

| HTTP | Причина |
|------|---------|
| 400 | Invalid composition / compatibility |
| 409 | GP version уже существует |
| 401 | Неверный admin key |

## coin-ui

GP publish wizard и audit log viewer — [coin-ui user guide](../coin-ui-user-guide.md) (`/releases/publish`, `/audit`).

Curl-примеры ниже — для automation и component publish.

## Связанные документы

- [golden-path-versioning.md](../golden-path-versioning.md)
- [control-plane.md](../control-plane.md)
- [coin-api/openapi/v1.yaml](../../coin-api/openapi/v1.yaml)
