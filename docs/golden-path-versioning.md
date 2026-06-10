# Версионирование Golden Path (Control Plane v2)

**Обновлено:** P2-08 — publish через Admin API.

> **Legacy v1** (`template` + `templateVersion`, `profile.yaml`) — DEPRECATED. См. [migrate-config-v1-to-v2.md](how-to/migrate-config-v1-to-v2.md).

## Модель v2

| Сущность | Semver | Где хранится | Кто pin'ит |
|----------|--------|--------------|------------|
| **GP release** | `go-app@1.0.0` | `gp_releases` + Nexus manifest | **Продукт** — 2 строки в config |
| **Component** | независимый cadence | `component_versions` | Platform — composition при publish |
| **Content** | PG + Nexus | scripts, Dockerfile, schema — seed: `coin-api/internal/gpcontent/seed/` |

Продукт:

```yaml
coin:
  goldenPath: go-app
  version: "1.0.0"
```

Всё остальное (executor, agent image, scripts, Dockerfile) — в **manifest**, собирается coin-api при Resolve.

## Lifecycle GP release

```
draft (snapshot) → published → deprecated → retired
```

| Статус | Resolve (product CI) | Nexus |
|--------|----------------------|-------|
| `draft` | ✅ только explicit pin `=1.0.0-snapshot.N` | ❌ |
| `published` | ✅ | blob + pointers |
| `deprecated` | ✅ + Warning header | ✅ |
| `retired` | ❌ | blob остаётся (immutable) |
| below `catalog.minimum` | ❌ 403 | — |

Draft/snapshot версии (`1.0.0-snapshot.N`) **не попадают** в catalog wildcards (`*`, `~`, `^`).

## Publish (append-only)

**Никогда не UPDATE** существующий GP version — только новый semver (`1.0.1`, `1.1.0`, `2.0.0`).

Flow:

1. Platform публикует/обновляет **components** (executor, pipeline, …)
2. (Optional) `POST .../drafts` — snapshot в PG для редактирования artifacts
3. `POST .../versions` или **promote draft** — composition + compatibility check
4. Resolve → canonical JSON + `manifestHash` → Nexus blob + mutable pointers
5. Audit log

Authoring: seed bytes в `coin-api/internal/gpcontent/seed/`, runtime — Nexus URLs в manifest.

How-to: [publish-gp-release.md](how-to/publish-gp-release.md)

## Catalog policy

Per GP в `catalog_policy`:

| Поле | Назначение |
|------|------------|
| `latest` | Stable line — pin `*`, Nexus pointer `latest.json` |
| `latest_canary` | Canary line — только API resolve с `project` |
| `minimum` | EOL enforcement на Resolve |
| `deprecated` | Warning, но Resolve OK |

Подробнее: [canary.md](canary.md)

## Compatibility matrix

При publish GP проверяется graph constraints (DB `component_compatibility`).

Пример: `pipeline/go-build@2.1.x` требует `executor >=0.1.0 <0.2.0`, `agent >=1.22.0`.

## Blast radius

Перед bump minimum или mass migration — `GET .../blast-radius` или coin-ui → GP Releases.

## Major vs minor GP

| Изменение | GP version |
|-----------|------------|
| Patch scripts / Dockerfile hash | `1.0.x` (новый component version + новый GP patch) |
| Новый agent runtime | часто `1.x.0` или `2.0.0` |
| Breaking config schema | **`2.0.0`** |

GP semver **≠** component semver — независимые cadence.

## Local pilot

- Только `go-app` в catalog
- E2E: `samples/demo-go-app@1.0.0`
- Corp fleet rollout — после corp gate

## Связанные документы

- [golden-paths.md](golden-paths.md)
- [control-plane.md](control-plane.md)
- [openapi.md](openapi.md)
