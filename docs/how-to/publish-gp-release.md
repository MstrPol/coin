# Publish GP release (Admin API)

**Audience:** Platform team  
**Scope:** local stend + подготовка к corp (без wave rollout).

## Prerequisites

- coin-api с Admin API
- Компоненты опубликованы в DB (`component_versions`)
- `COIN_ADMIN_API_KEY` или `AUTH_DISABLED=true` (local)

## Composition (4 slots)

Для `go-app` (и аналогов `go-app-bp`, `go-app-df`):

| Key | Component type | Пример |
|-----|----------------|--------|
| `agent` | `agent` | `coin-agent@1.0.0` |
| `executor` | `executor` | `coin-executor@0.1.0` |
| `lib` | `lib` | `coin-lib@1.0.0` |
| `gp-content` | `gp-content` | `gp-content/go-app@1.0.2` |

**Superseded:** slots `pipeline`, `validate`, `dockerfile`, stack agent `go@{ver}`.

Manifest собирается из composition: `build`, `pipeline.stages` (typed), `validateSchema`, `runtime.image`.

## Local seed (рекомендуется)

```bash
cd docker
make publish-agent GOARCH=arm64   # при необходимости
make coin-lib
make seed-jenkins-lib             # публикует components + GP go-app / go-app-bp / go-app-df
```

Скрипт: `docker/scripts/seed-jenkins-lib-stack.sh`.

## Publish GP (Admin API)

```bash
curl -X POST http://localhost:8090/v1/admin/golden-paths/go-app/versions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${COIN_ADMIN_API_KEY:-dev-local-admin-key}" \
  -d '{
    "version": "1.0.3",
    "composition": {
      "agent": "1.0.0",
      "executor": "0.1.0",
      "lib": "1.0.0",
      "gp-content": "1.0.2"
    },
    "actor": "platform-team"
  }'
```

**Side-effects:**

1. Row в `gp_releases` + `gp_composition` (append-only)
2. Resolve → canonical manifest → Nexus blob + pointers
3. Audit log

**Важно:** Nexus blobs immutable — при смене gp-content sha256 нужен **новый** gp-content version и новый GP release, не UPDATE blob.

### Draft / snapshot (optional)

```bash
curl -X POST http://localhost:8090/v1/admin/golden-paths/go-app/drafts \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${COIN_ADMIN_API_KEY:-dev-local-publisher-key}" \
  -d '{"version":"1.0.3-snapshot.1","baseVersion":"1.0.2","actor":"platform-team"}'

curl -X POST "http://localhost:8090/v1/admin/golden-paths/go-app/versions/1.0.3-snapshot.1/promote" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${COIN_ADMIN_API_KEY:-dev-local-publisher-key}" \
  -d '{"actor":"platform-team"}'
```

coin-ui: [Publish wizard](../coin-ui-user-guide.md).

### Verify

```bash
curl -sf http://localhost:8090/v1/golden-paths/go-app/versions/1.0.3/manifest \
  | jq '{hash: .manifestHash, engine: .build.engine, runtime: .runtime.image}'

SNAPSHOTS=http://localhost:8081/repository/maven-snapshots
curl -sf "${SNAPSHOTS}/coin/manifest/go-app/metadata/go-app-metadata-pin-%3D1.0.3.json" | jq '{manifestHash, blobUrl}'
```

### Publish component

```bash
curl -X POST http://localhost:8090/v1/admin/components/gp-content/go-app/versions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${COIN_ADMIN_API_KEY}" \
  -d '{"version":"1.0.3","metadata":{}}'
```

Затем `POST .../golden-paths/go-app/versions` с обновлённым `gp-content` pin.

Duplicate version → **409**.

## Ошибки

| HTTP | Причина |
|------|---------|
| 400 | Invalid composition / compatibility |
| 409 | GP version уже существует |
| 401 | Неверный admin key |

## Связанные документы

- [golden-path-versioning.md](../golden-path-versioning.md)
- [golden-paths.md](../golden-paths.md)
- [agent-build-model.md](../agent-build-model.md)
- [coin-api/openapi/v1.yaml](../../coin-api/openapi/v1.yaml)
