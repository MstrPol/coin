# Control Plane v2

Три слоя источника правды, **platform components** (UI-first) и runtime на `coin-executor`.

## Три слоя

| Слой | Где | Что хранит |
|------|-----|------------|
| **Content** | Nexus (+ PG draft bodies в Studio) | gp-content packages, executor binary (Nexus bake only), manifest blobs; coin-lib ZIP — вне PG registry |
| **Metadata** | PostgreSQL | `component_versions`, GP releases, composition, catalog policy, audit |
| **Runtime cache** | Nexus `maven-releases` / `maven-snapshots` | immutable manifest blobs + mutable pointers |

**Q1 (platform lead):** `component_artifact_bodies` — только **draft** в Component Studio. Published/canary — **Nexus-only** (без dual-write тел в PG).

Manifest — **канонический JSON** с `manifestHash` (sha256). Собирается coin-api при Resolve из composition slots + materializers и кешируется в Nexus.

Миграция legacy dual-write (`gp_artifact_bodies`): [runbooks/gp-artifact-bodies-migration.md](runbooks/gp-artifact-bodies-migration.md).

## Platform components (UI-first)

Enabling team выпускает platform content через **coin-ui Platform catalogs** → Admin API → Nexus → registry.

| Роль | Путь |
|------|------|
| **Primary** | Platform editor (`/platform/build-stacks/.../edit`, `/platform/branching-models/.../edit`) → validate → register → promote (`draft`→`published`) |
| **GP pin** | coin-ui GP hub — composition в `gp_releases` |
| **Promote** | GP draft→published с gate: все pins `published`; catalog `latest_canary`→`latest` — отдельный wizard |
| **Deprecated** | `/studio`, `publish-canary`, `publish-*.sh`, Gitea platform jobs |

### Lifecycle (`component_versions`)

| State | Product resolve (stable) | Canary channel (GP) | Platform edit |
|-------|--------------------------|---------------------|---------------|
| `draft` | ❌ | ✅ draft pins на canary line | ✅ |
| `published` | ✅ | ✅ | ❌ |

Переход: `draft` → validate → register (PG) → **promote** (Nexus) → `published`.

**Canary** — только на уровне GP (`catalog.latest_canary`); может указывать на GP `draft`. См. [canary.md](canary.md).

ADR: [adr/gp-component-package-model.md](adr/gp-component-package-model.md) (amended 2026-06).

## Компоненты

| Компонент | Роль |
|-----------|------|
| **coin-api** | Resolve manifest, registry, GP admin, component lifecycle, build report; bootstrap seed `internal/gpcontent/seed/` |
| **coin-executor** | `validate`, `run --stage`, `publish`, `report` |
| **coin-lib** | Jenkins glue (resolve, pod, stages); ZIP из Nexus HTTP |
| **coin-ui** | Admin SPA + Platform catalogs + promote wizard |

## Manifest (v1 hard cut, сокращённо)

`manifestVersion` остаётся `1`: local pilot использует hard cut контракта, поэтому superseded поля удаляются без compatibility shim. Build Stack vNext описан в [adr/build-stack-vnext-contract.md](adr/build-stack-vnext-contract.md).

```json
{
  "manifestVersion": 1,
  "manifestHash": "sha256:…",
  "goldenPath": { "name": "go-app", "version": "1.0.2" },
  "runtime": {
    "image": "nexus:8082/coin-docker/coin-agent:1.0.0",
    "digest": "sha256:…"
  },
  "destinations": {
    "imageRegistryPrefix": "docker-dev.registry.domain.ru",
    "buildCacheEnabled": true,
    "artifactRepositoryBase": "http://nexus:8081/repository/maven-releases"
  },
  "parameters": [
    { "name": "GO_VERSION", "type": "string", "default": "1.22", "required": true }
  ],
  "build": {
    "targets": [
      { "id": "app-image", "engine": "buildkit", "containerfile": "app", "target": "runtime" },
      { "id": "app-artifact", "engine": "buildkit", "containerfile": "app", "target": "artifact" }
    ]
  },
  "deliverables": [
    { "id": "app", "type": "image", "targetId": "app-image" },
    { "id": "app-zip", "type": "artifact", "targetId": "app-artifact" }
  ],
  "artifacts": {
    "containerfiles": [
      { "id": "app", "url": "…", "sha256": "sha256:…" }
    ]
  },
  "validateSchema": {
    "url": "http://nexus:8081/repository/maven-releases/coin/gp/content/go-app/1.0.2/config.v2.schema.json",
    "sha256": "sha256:…"
  },
  "branching": {
    "name": "trunk-based",
    "version": "1.0.0",
    "branches": [
      { "name": "main", "pattern": "main", "versioning": { "template": "0.0.0-SNAPSHOT" }, "publish": true }
    ]
  },
  "pipeline": {
    "stages": [
      { "id": "validate", "name": "Validate", "steps": [{ "action": "run-target", "targetId": "app-image" }] },
      { "id": "build", "name": "Build", "steps": [{ "action": "build-deliverable", "deliverableId": "app" }] },
      { "id": "publish", "name": "Publish", "steps": [{ "action": "publish-deliverable", "deliverableId": "app" }] }
    ]
  }
}
```

Product repo не задаёт build/publish outputs. Deliverables являются частью GP/Build Stack policy и материализуются в manifest как named outputs; в P0 разрешены несколько deliverables одного type.

Stage `publish`: coin-lib skip при `params.publish=false`; eligibility — `manifest.branching` + `COIN_PUBLISH_REQUEST`. См. [adr/gp-branching-model.md](adr/gp-branching-model.md).

**Superseded в manifest:** `dockerfileTemplate`, `pipeline.stages[].script`, `manifest.jnlp`, orchestration bundle URL, `pipeline.stages[].when` как primary publish gate, `lib`, `executor`, Jenkins `credentials`.

OpenAPI: [`coin-api/openapi/v1.yaml`](../coin-api/openapi/v1.yaml).  
Schema: [`coin-api/manifest.schema.json`](../coin-api/manifest.schema.json).

## Resolve materializers

coin-api собирает manifest через **composition slot registry** (не switch per type):

1. `gp_composition` → pin component type/name/version per slot
2. Materializer загружает package / metadata (`content_ref` v2 или legacy)
3. `manifest.Builder` денормализует GP identity + destinations версии GP + секции composition pins (`runtime`, `build`, `pipeline`, `validateSchema`, `capabilities`, `branching`)

CI fallback при недоступном API — **только Nexus** (manifest blob + component packages), не PG bodies.

## CI flow

```mermaid
sequenceDiagram
  participant J as Jenkins (coin-lib)
  participant API as coin-api
  participant NX as Nexus
  participant POD as coin-agent pod
  participant EX as coin-executor

  J->>J: read .coin/config.yaml
  J->>API: GET resolve (primary)
  alt API down
    J->>NX: pointer → manifest blob (verify manifestHash)
  end
  J->>POD: podTemplate(runtime.image)
  POD->>POD: podman system service
  POD->>EX: run --stage validate
  POD->>EX: run --stage test
  POD->>EX: run --stage build
  POD->>EX: run --stage publish (optional)
  POD->>EX: report → API
```

Build dispatch идёт по `manifest.build.targets[].engine` — см. [adr/build-stack-vnext-contract.md](adr/build-stack-vnext-contract.md) и [adr/coin-ci-runtime.md](adr/coin-ci-runtime.md).

## Миграция с v1

Config v1 (`template`/`templateVersion`, fat pipeline, Shared Library business logic) **выведен**. См. [migrate-config-v1-to-v2.md](how-to/migrate-config-v1-to-v2.md).

ADR: [`docs/adr/control-plane-v2.md`](adr/control-plane-v2.md).
