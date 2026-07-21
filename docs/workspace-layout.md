# Workspace layout (local pilot + corp split)

Канон расположения кода для local pilot и подготовка к corp extract. OpenSpec: `docs-monorepo-layout`.

## Два уровня

| Уровень | Что это | Когда |
|---------|---------|--------|
| **Integration workspace** | Sibling-каталоги на машине / CI checkout | **сейчас** (local pilot) |
| **Corp prod repos** | Отдельные Gitea `coin/<name>` | **после corp gate** — [prod-repo-split](runbooks/prod-repo-split.md) |

Local compose **не** требует corp split.

## Диаграмма (сейчас)

```
coin-workspace/                          # интеграционный workspace
├── coin-api/                            # control plane API + PG migrations + seed
├── coin-executor/                       # CLI runtime + Dockerfile.agent (coin-agent)
├── coin-lib/                            # Jenkins Shared Library (glue only)
├── coin-ui/                             # Admin / Platform SPA
└── coin/                                # meta-репозиторий
    ├── docs/                            # эта документация
    ├── openspec/                        # specs + changes (канон требований)
    ├── docker/                          # compose: Gitea, Jenkins, Nexus, k3s, …
    ├── coin-starters/                   # scaffolding продуктовых репо
    └── samples/                         # E2E product demos (локальные клоны)
```

Sibling-репозитории **не** лежат внутри `coin/` — они рядом с ним в workspace.

## Inventory

| Path | Роль | Local | Будущий corp repo |
|------|------|-------|-------------------|
| `coin-api/` | Resolve, Admin API, registry, GP, seed embed | compose `coin-api` | `coin/coin-api` |
| `coin-executor/` | `validate` / `run` / `publish` / `report`; agent image | bake в `publish-agent` | `coin/coin-executor` |
| `coin-lib/` | Jenkins `@Library` glue | `make coin-lib` / HTTP ZIP | `coin/coin-lib` |
| `coin-ui/` | Platform UI | compose `coin-ui` | `coin/coin-ui` |
| `coin/docs`, `openspec` | Docs + OpenSpec | — | обычно остаётся platform docs repo / monorepo meta |
| `coin/docker` | Local pilot stack | `make bootstrap` | **не** prod |
| `coin/coin-starters` | Product scaffolding | `make coin-starters` | `coin/coin-starters` |
| `coin/samples` | E2E demos | `make samples` → Gitea | опционально `coin/samples` |

### Samples (Q1)

**Канон local pilot:** `coin/samples/` — пишет `docker/scripts/samples.sh` (`SAMPLES_DIR="${REPO_ROOT}/samples"`, `REPO_ROOT` = `coin/`).

Каталог `samples/` на корне workspace (рядом с `coin/`), если есть, **не** SoT для скриптов docker; не путать с `coin/samples`.

## Seed SoT (не отдельные package-репо)

| Что | Где |
|-----|-----|
| Pipeline defaults (`go-app`, `go-app-docker`) | `coin-api/internal/gpcontent/seed/` |
| Branching-model fixtures (seed/E2E) | `coin/docker/testdata/branching-models/` |
| Live authoring pipeline | GP release в Platform UI |
| Live authoring branching | `/platform/branching-models` |

## Удалённые / superseded деревья

| Было | Статус | Замена |
|------|--------|--------|
| `coin/coin-gp-content/` | **удалено** | api seed + embedded GP pipeline |
| `coin/coin-branching-models/` | **удалено** | Platform + `docker/testdata/branching-models/` |
| `coin-jenkins-agents/` | **superseded** | universal `coin-agent` (`coin-executor/Dockerfile.agent`) |

## GP composition (напоминание)

Два pin: `agent` + `branching-model`. Pipeline — **embedded** на GP release. См. [architecture](architecture.md), [gp-embedded-pipeline](adr/gp-embedded-pipeline.md).

## Corp split

Подготовка и checklist — [runbooks/prod-repo-split.md](runbooks/prod-repo-split.md) (P4-03, corp gate). На local pilot **не выполнять** `git filter-repo`.

## См. также

- [architecture.md](architecture.md)
- [openspec/specs/docs-monorepo-layout](../openspec/specs/docs-monorepo-layout/spec.md)
- [docker/README.md](../docker/README.md)
