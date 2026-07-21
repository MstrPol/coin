# Publish GP release (Admin API)

**Audience:** Platform team  
**Scope:** local stend + подготовка к corp (без wave rollout).

## Prerequisites

- coin-api с Admin API
- Agent и branching-model опубликованы в DB (`component_versions`)
- `COIN_ADMIN_API_KEY` или `AUTH_DISABLED=true` (local)

## Composition (2 pins)

Для `go-app`, `go-app-docker` и других GP profiles:

| Key | Component type | Пример |
|-----|----------------|--------|
| `agent` | `agent` | `coin-agent@1.0.0` |
| `branching-model` | `branching-model` | `trunk-based@1.0.0` |

Pipeline-inline v3 (`parameters`, `pipeline.stages`, containerfiles) — **embedded body** на GP release draft, не отдельный component pin.

Agent pin = полный CI runtime stack (`coin-executor` baked в образ). `coin-lib` — вне GP composition (Jenkins `@Library`).

**Superseded:** 3-pin с `gp-content`; platform component type `gp-content`.

Manifest собирается из embedded pipeline + composition pins: `pipeline.stages`, `validateSchema`, `branching`, `runtime.image`.

См. [adr/gp-embedded-pipeline.md](../adr/gp-embedded-pipeline.md), [adr/coin-ci-runtime.md](../adr/coin-ci-runtime.md).

## Local seed (рекомендуется)

```bash
cd docker
make publish-agent GOARCH=arm64   # при необходимости
make seed-jenkins-lib             # Nexus lib + branching-model + GP (2-pin + embedded pipeline)
```

Скрипт: `docker/scripts/seed-jenkins-lib-stack.sh`.  
**Deprecated:** `make coin-lib` (Gitea SCM only); бывший `coin-gp-content/scripts/publish-content.sh` (папка удалена).

## Publish GP (Admin API)

```bash
curl -X POST http://localhost:8090/v1/admin/golden-paths/go-app/versions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${COIN_ADMIN_API_KEY:-dev-local-admin-key}" \
  -d '{
    "version": "1.0.3",
    "destinations": {
      "imageRegistryPrefix": "localhost:8082/coin-docker",
      "buildCacheEnabled": true,
      "artifactRepositoryBase": "http://nexus:8081/repository/maven-releases"
    },
    "agentStackName": "coin-agent",
    "branchingModelName": "trunk-based",
    "composition": {
      "agent": "1.0.0",
      "branching-model": "1.0.0"
    },
    "actor": "platform-team"
  }'
```

При первом publish pipeline body seed'ится из embedded defaults (`coin-api/internal/gpcontent/seed/pipelines/`).

**Side-effects:**

1. Row в `gp_releases` + `gp_composition` + `gp_release_pipeline_bodies`
2. Resolve → canonical manifest → Nexus blob + pointers
3. Audit log

**Важно:** Nexus blobs immutable — при смене pipeline нужен **новый** GP release version, не UPDATE blob.

### Draft / promote (recommended)

```bash
curl -X POST http://localhost:8090/v1/admin/golden-paths/go-app/drafts \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${COIN_ADMIN_API_KEY:-dev-local-publisher-key}" \
  -d '{
    "version": "1.0.3-snapshot.1",
    "destinations": {
      "imageRegistryPrefix": "localhost:8082/coin-docker",
      "buildCacheEnabled": true,
      "artifactRepositoryBase": "http://nexus:8081/repository/maven-releases"
    },
    "agentStackName": "coin-agent",
    "branchingModelName": "trunk-based",
    "composition": {"agent": "1.0.0", "branching-model": "1.0.0"},
    "actor": "platform-team"
  }'

# Редактировать pipeline: PUT .../versions/{version}/pipeline

curl -X POST "http://localhost:8090/v1/admin/golden-paths/go-app/versions/1.0.3-snapshot.1/promote" \
  -H "X-API-Key: ${COIN_ADMIN_API_KEY:-dev-local-publisher-key}"
```

coin-ui: GP hub → release detail → Pipeline section.

### Verify

```bash
curl -sf http://localhost:8090/v1/golden-paths/go-app/versions/1.0.3/manifest \
  | jq '{hash: .manifestHash, stages: (.pipeline.stages | length), runtime: .runtime.image}'

SNAPSHOTS=http://localhost:8081/repository/maven-snapshots
curl -sf "${SNAPSHOTS}/coin/manifest/go-app/metadata/go-app-metadata-pin-%3D1.0.3.json" | jq '{manifestHash, blobUrl}'
```

Duplicate version → **409**.

## Ошибки

| HTTP | Причина |
|------|---------|
| 400 | Invalid composition / pipeline validation |
| 409 | GP version уже существует; promote blocked by draft pins |
| 401 | Неверный admin key |

## Связанные документы

- [golden-path-versioning.md](../golden-path-versioning.md)
- [golden-paths.md](../golden-paths.md)
- [agent-build-model.md](../agent-build-model.md)
- [coin-api/openapi/v1.yaml](../../coin-api/openapi/v1.yaml)
