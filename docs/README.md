# Документация Coin CI (Control Plane v2)

## Быстрый старт

| Документ | Для кого |
|----------|----------|
| **[how-to/onboarding-15min.md](how-to/onboarding-15min.md)** | **Новый dev — стенд за 15 min** |
| [how-to/local-dev-control-plane.md](how-to/local-dev-control-plane.md) | Platform — поднять стенд, resolve manifest |
| [how-to/add-new-service-repo.md](how-to/add-new-service-repo.md) | Команда — новый сервис на go-app |
| [how-to/migrate-config-v1-to-v2.md](how-to/migrate-config-v1-to-v2.md) | Команда — миграция с config v1 |
| [how-to/p0-go-no-go-checklist.md](how-to/p0-go-no-go-checklist.md) | Lead — gate фазы 0 |
| [how-to/wave-migration-checklist.md](how-to/wave-migration-checklist.md) | PM / owner — миграция одного repo |
| [how-to/fleet-analytics-pm.md](how-to/fleet-analytics-pm.md) | PM — blast radius, adoption, stale |
| [how-to/troubleshoot-ci.md](how-to/troubleshoot-ci.md) | On-call — ошибки CI v2 |
| [runbooks/wave-1-migration.md](runbooks/wave-1-migration.md) | PM — Wave 1 (50 repos) |
| [runbooks/wave-3-migration.md](runbooks/wave-3-migration.md) | PM — Wave 3 (1500+) + comms templates |
| [runbooks/scanner-ops.md](runbooks/scanner-ops.md) | ~~fleet scanner~~ (superseded — build reports) |
| [how-to/publish-gp-release.md](how-to/publish-gp-release.md) | Platform — publish GP через Admin API |
| [runbooks/gp-artifact-bodies-migration.md](runbooks/gp-artifact-bodies-migration.md) | Platform — dual-write cleanup plan (GCP-5) |
| [coin-ui-user-guide.md](coin-ui-user-guide.md) | PM — dashboard coin-ui |
| [openapi.md](openapi.md) | OpenAPI / Swagger контракт |
| [runbooks/api-down-nexus-fallback.md](runbooks/api-down-nexus-fallback.md) | On-call — API down |
| [runbooks/prod-repo-split.md](runbooks/prod-repo-split.md) | Platform — P4-03 corp split |
| [docker/README.md](../docker/README.md) | Docker Compose стенд |

## Архитектура

| Документ | Содержание |
|----------|------------|
| [architecture.md](architecture.md) | Обзор компонентов v2 |
| [control-plane.md](control-plane.md) | API, executor, manifest, три слоя SoT |
| [agent-build-model.md](agent-build-model.md) | **Build engines** runbook (E2E, troubleshooting) |
| [adr/coin-ci-runtime.md](adr/coin-ci-runtime.md) | **CI runtime** canonical ADR |
| [golden-paths.md](golden-paths.md) | GP profiles, three-pin composition, samples |
| [responsibilities.md](responsibilities.md) | Platform vs команда |
| [planning.md](planning.md) | **OpenSpec + Beads** — workflow, активные changes |

## Контракты

| Документ | Содержание |
|----------|------------|
| [config.md](config.md) | `.coin/config.yaml` v2 (`goldenPath` + `version`) |
| [jenkins-setup.md](jenkins-setup.md) | coin-lib, K8s pod, platform jobs |
| [how-to/branching-models.md](how-to/branching-models.md) | Ветки, теги, branching-model v2 |

## Компоненты monorepo

| Документ | Содержание |
|----------|------------|
| [coin-api/README.md](../coin-api/README.md) | Resolve API, admin, migrations |
| [coin-ui/README.md](../coin-ui/README.md) | Admin SPA (local :8091) |
| [coin-executor/README.md](../coin-executor/README.md) | CLI runtime, coin-agent, build engines |
| [coin-lib/README.md](../coin-lib/README.md) | Jenkins Shared Library (glue) |
| [coin-starters/README.md](../coin-starters/README.md) | Product scaffolding |
| [how-to/publish-gp-release.md](how-to/publish-gp-release.md) | GP release + embedded pipeline |
| [adr/gp-embedded-pipeline.md](adr/gp-embedded-pipeline.md) | Seed: coin-api `internal/gpcontent/seed/` |

## Прочее

| Документ | Содержание |
|----------|------------|
| [golden-path-versioning.md](golden-path-versioning.md) | GP semver + publish model (v2) |
| [release-notes.md](release-notes.md) | QGM, smart-коммиты |

## Ключевые принципы

1. **Продукт** задаёт только `coin.goldenPath` + `coin.version` — build engine и stages в manifest.
2. **coin-api** собирает manifest из composition slots + Nexus packages (materializers).
3. **Component Studio** — primary path для platform components (`gp-content`, `branching-model`).
4. **Nexus** — immutable blobs + mutable pointers; CI resolve с fallback при недоступном API.
5. **coin-executor** — validate, build engines (`buildkit` / BYO `dockerfile`), report.
6. **coin-lib** + **Jenkinsfile.coin** — resolve → pod (`coin-agent`) → executor stages.
7. **E2E local pilot:** `make e2e-build-engines` — две demo jobs (buildkit + BYO dockerfile).

## ADR (архитектурные решения)

| ADR | Тема |
|-----|------|
| [coin-ci-runtime](adr/coin-ci-runtime.md) | **CI runtime** (agent, bootstrap, engines, publish) |
| [build-engine-contract](adr/build-engine-contract.md) | `build.engine` hard cut decision |
| [gp-component-package-model](adr/gp-component-package-model.md) | UI-first components, package model |
| [gp-branching-model](adr/gp-branching-model.md) | branching-model component, publish policy |
| [jenkins-lib-http-nexus](adr/jenkins-lib-http-nexus.md) | coin-lib scope |
| [control-plane-v2](adr/control-plane-v2.md) | manifest v2 |

Планирование и backlog: [planning.md](planning.md).

## Doc review

| Check | Status |
|-------|--------|
| Build engine model (3 samples, E2E) | ✅ |
| Superseded: coin-jenkins-agents, script URLs | ✅ |
| P4-03 prod Gitea split | ⏸ corp gate |
