# Publish GP release (Admin API)

**Audience:** Platform team  
**Scope:** local stend + подготовка к corp (без wave rollout).

## Prerequisites

- coin-api с Admin API
- Компоненты опубликованы в DB (`component_versions`)
- `COIN_ADMIN_API_KEY` или `AUTH_DISABLED=true` (local)

## Composition (3 pins)

Для `go-app` (и аналогов `go-app-bp`, `go-app-df`):

| Key | Component type | Пример |
|-----|----------------|--------|
| `agent` | `agent` | `coin-agent@1.0.0` |
| `gp-content` | `gp-content` | `go-app@1.0.2` |
| `branching-model` | `branching-model` | `trunk-based@1.0.0` |

Agent pin = полный CI runtime stack (`coin-executor` baked в образ). `coin-lib` — вне GP composition (Jenkins `@Library`).

**Superseded:** 4-slot с `executor` + `lib` в composition; `manifest.executor` секция.

Manifest собирается из composition: `build`, `pipeline.stages` (typed), `validateSchema`, `branching`, `runtime.image`.

См. [adr/coin-ci-runtime.md](../adr/coin-ci-runtime.md).

## Local seed (рекомендуется)

```bash
cd docker
make publish-agent GOARCH=arm64   # при необходимости
make seed-jenkins-lib             # Nexus lib ZIP + gp-content + GP + coin-lib-http
```

Скрипт: `docker/scripts/seed-jenkins-lib-stack.sh`.  
**Deprecated:** `make coin-lib` (Gitea SCM only).

## Publish GP (Admin API)

```bash
curl -X POST http://localhost:8090/v1/admin/golden-paths/go-app/versions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${COIN_ADMIN_API_KEY:-dev-local-admin-key}" \
  -d '{
    "version": "1.0.3",
    "composition": {
      "agent": "1.0.0",
      "gp-content": "1.0.2",
      "branching-model": "1.0.0"
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
