# Golden Paths

> **Канон:** composition — **2 pin** (`agent` + `branching-model`); pipeline-inline — **embedded body** GP release.  
> Specs: `gp-release-two-pin`, `gp-embedded-pipeline`, `gp-entity-hub`.  
> ADR: [gp-embedded-pipeline.md](adr/gp-embedded-pipeline.md).

## Термины

| Термин | Смысл |
|--------|--------|
| **GP profile** | Имя семейства (`go-app`, `go-app-docker`) = `coin.goldenPath` |
| **GP release** | Версия профиля: composition pins + embedded pipeline + destinations |
| **Platform component** | `agent`, `branching-model` (semver, draft→published) |

## Authoring flow

| Шаг | Где |
|-----|-----|
| 1. Agent published | `/platform/runtime` |
| 2. Branching-model published (или draft pin на canary) | `/platform/branching-models` |
| 3. GP draft | `/gp/{name}` → new draft: pin agent + branching |
| 4. Pipeline | Release detail → Pipeline editor (embedded) |
| 5. Promote | Gate: pins `published` + pipeline valid → `published` |

**Local bootstrap:** `make seed-jenkins-lib` — lib + branching fixtures + GP profiles с pipeline из coin-api seed.

**Deprecated:** `/studio`, `publish-content.sh`, папка `coin-gp-content/`, 3-pin с `gp-content`.

## Composition (2 pin)

```
GP release
├── pin agent ──────────────► manifest.runtime
├── pin branching-model ────► manifest.branching
└── embedded pipeline ──────► manifest.pipeline (+ related build fragments)
```

| Key | Type | Пример |
|-----|------|--------|
| `agent` | `agent` | `coin-agent@1.0.0` |
| `branching-model` | `branching-model` | `trunk-based@1.0.0` |

Agent pin в composition — только `published`. Branching на draft GP / canary channel может быть `draft`.

## Bootstrap seed

```
coin-api/internal/gpcontent/seed/pipelines/
├── go-app.yaml
└── go-app-docker.yaml
```

Branching fixtures: `coin/docker/testdata/branching-models/`.  
Seed script: `docker/scripts/seed-branching-model.sh` + `seed-jenkins-lib-stack.sh`.

## Build engines (local pilot)

| GP | Engine | Sample |
|----|--------|--------|
| `go-app` | buildkit | `coin/samples/demo-go-app` |
| `go-app-docker` | dockerfile (BYO) | `coin/samples/demo-go-app-docker` |

## Publish eligibility

Jenkins `params.publish` → `COIN_PUBLISH_REQUEST` + `manifest.branching` rules.  
Не primary gate: `pipeline.stages[].when: tag`.

## Canary

Catalog `latest_canary`, `project.canary_mode`, `X-Coin-Channel`. См. [canary.md](canary.md).

## См. также

- [how-to/publish-gp-release.md](how-to/publish-gp-release.md)
- [how-to/branching-models.md](how-to/branching-models.md)
- [architecture.md](architecture.md)
- [workspace-layout.md](workspace-layout.md)
