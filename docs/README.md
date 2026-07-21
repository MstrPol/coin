# Документация Coin CI (Control Plane v2)

Канон требований: [`openspec/specs/`](../openspec/specs/).  
Расположение кода: **[workspace-layout.md](workspace-layout.md)**.

## Быстрый старт

| Документ | Для кого |
|----------|----------|
| **[how-to/onboarding-15min.md](how-to/onboarding-15min.md)** | Новый dev — стенд |
| [how-to/local-dev-control-plane.md](how-to/local-dev-control-plane.md) | Platform — стенд, resolve |
| [how-to/add-new-service-repo.md](how-to/add-new-service-repo.md) | Команда — новый сервис |
| [how-to/publish-gp-release.md](how-to/publish-gp-release.md) | Platform — GP + embedded pipeline |
| [how-to/branching-models.md](how-to/branching-models.md) | Platform — branching-model |
| [how-to/troubleshoot-ci.md](how-to/troubleshoot-ci.md) | On-call |
| [coin-ui-user-guide.md](coin-ui-user-guide.md) | Operators / PM |
| [runbooks/prod-repo-split.md](runbooks/prod-repo-split.md) | Corp P4-03 (после corp gate) |
| [runbooks/api-down-nexus-fallback.md](runbooks/api-down-nexus-fallback.md) | On-call — API down |
| [docker/README.md](../docker/README.md) | Compose стенд |

## Архитектура и layout

| Документ | Содержание |
|----------|------------|
| **[workspace-layout.md](workspace-layout.md)** | Sibling repos + `coin/` meta + removed trees |
| [architecture.md](architecture.md) | Компоненты, **2-pin** composition |
| [control-plane.md](control-plane.md) | SoT layers, Platform hubs, resolve |
| [golden-paths.md](golden-paths.md) | GP profiles, authoring, seed |
| [agent-build-model.md](agent-build-model.md) | Build engines runbook |
| [adr/coin-ci-runtime.md](adr/coin-ci-runtime.md) | CI runtime ADR |
| [adr/gp-embedded-pipeline.md](adr/gp-embedded-pipeline.md) | Embedded pipeline ADR |
| [responsibilities.md](responsibilities.md) | Platform vs команда |
| [planning.md](planning.md) | OpenSpec + Beads |

## Контракты

| Документ | Содержание |
|----------|------------|
| [config.md](config.md) | Product `.coin/config.yaml` v2 |
| [openapi.md](openapi.md) | OpenAPI / Swagger |
| [schemas/branching-model.schema.json](schemas/branching-model.schema.json) | Branching model schema docs |

## OpenSpec (обязательный канон)

| Capability | Тема |
|------------|------|
| `gp-release-two-pin` | Composition 2 pin |
| `gp-embedded-pipeline` | Pipeline на GP release |
| `docs-monorepo-layout` | Workspace / corp split |
| `runtime-documentation` | Docs ↔ ADR consistency |
| `jenkins-lib-boundary` | coin-lib glue only |

Полный список: [`openspec/specs/`](../openspec/specs/).

## Superseded how-to

| Документ | Статус |
|----------|--------|
| [how-to/build-stacks.md](how-to/build-stacks.md) | Redirect → GP embedded pipeline |

## Reading order (platform)

1. [workspace-layout.md](workspace-layout.md)
2. [architecture.md](architecture.md) + [control-plane.md](control-plane.md)
3. [golden-paths.md](golden-paths.md) + [how-to/publish-gp-release.md](how-to/publish-gp-release.md)
4. [adr/coin-ci-runtime.md](adr/coin-ci-runtime.md) + [agent-build-model.md](agent-build-model.md)
5. `openspec/specs/gp-release-two-pin` + `gp-embedded-pipeline`
