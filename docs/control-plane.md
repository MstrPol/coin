# Control Plane v2

Три слоя SoT, Platform UI-first authoring, runtime на `coin-executor`.  
Канон: `openspec/specs/` (`gp-release-two-pin`, `gp-embedded-pipeline`, `platform-*`). Layout: [workspace-layout.md](workspace-layout.md).

## Три слоя

| Слой | Где | Что хранит |
|------|-----|------------|
| **Content** | Nexus + PG drafts | published packages (`agent`, `branching-model`), GP manifest blobs; pipeline draft bodies в PG; coin-lib ZIP вне GP registry |
| **Metadata** | PostgreSQL | `component_versions`, GP releases, composition (2 pins), catalog policy, audit |
| **Runtime cache** | Nexus `maven-releases` / `maven-snapshots` | immutable manifest blobs + mutable pointers |

Resolve **не** требует live DB на product build path при наличии Nexus cache (primary HTTP coin-api, fallback Nexus).

## Platform authoring (UI-first)

| Роль | Путь |
|------|------|
| **Runtime (agent)** | `/platform/runtime/...` → draft → validate → publish |
| **Branching models** | `/platform/branching-models/...` → draft → validate → publish |
| **GP + pipeline** | `/gp/...` hub → release detail: composition (2 pins) + **Pipeline** editor (embedded) |
| **Promote GP** | draft→published; gate: pins `published` + valid embedded pipeline |
| **Deprecated** | `/studio`, Component Studio primary path, `publish-content.sh`, папки `coin-gp-content/` / `coin-branching-models/` |

### Lifecycle platform components (`agent`, `branching-model`)

| State | Product resolve (stable) | Draft / canary GP | Platform edit |
|-------|--------------------------|-------------------|---------------|
| `draft` | ❌ | ✅ (branching; agent — только `published` в composition) | ✅ |
| `published` | ✅ | ✅ | ❌ |

**Canary** — на уровне GP catalog (`latest_canary`), не component-level canary. См. [canary.md](canary.md).

## Компоненты

| Компонент | Роль |
|-----------|------|
| **coin-api** | Resolve, registry, GP admin; seed `internal/gpcontent/seed/` |
| **coin-executor** | `validate`, `run`, `publish`, `report` |
| **coin-lib** | Jenkins glue; ZIP из Nexus HTTP |
| **coin-ui** | Platform catalogs + GP hub |

## GP composition

Ровно два pin: `agent`, `branching-model`. Pipeline — embedded body. См. [architecture.md](architecture.md), [golden-paths.md](golden-paths.md).

## Manifest (v1)

Канонический JSON с `manifestHash` (sha256). Собирается coin-api при resolve/promote из composition + embedded pipeline + destinations.

Ключевые секции: `runtime`, `branching`, `pipeline` / build fragments, `destinations`, `validateSchema` (по schema).

**Нет в manifest:** `executor`, `lib` как composition materialization, orchestration bundle URL.

Stage `publish`: coin-lib skip при `params.publish=false`; eligibility — `manifest.branching` + `COIN_PUBLISH_REQUEST`. См. [adr/gp-branching-model.md](adr/gp-branching-model.md).

OpenAPI: sibling `coin-api/openapi/v1.yaml`.  
Schema: sibling `coin-api/manifest.schema.json`.

## Resolve

Primary: HTTP coin-api. Fallback: Nexus manifest blob.  
Materializers: agent → `runtime`; branching-model → `branching`; embedded pipeline → pipeline/build sections.

## См. также

- [adr/control-plane-v2.md](adr/control-plane-v2.md)
- [adr/gp-embedded-pipeline.md](adr/gp-embedded-pipeline.md)
- [how-to/publish-gp-release.md](how-to/publish-gp-release.md)
- [runbooks/api-down-nexus-fallback.md](runbooks/api-down-nexus-fallback.md)
