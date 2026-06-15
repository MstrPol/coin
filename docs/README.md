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
| [agent-build-model.md](agent-build-model.md) | Native build + runtime-only Dockerfile |
| [golden-paths.md](golden-paths.md) | GP, content/, composition |
| [responsibilities.md](responsibilities.md) | Platform vs команда |

## Контракты

| Документ | Содержание |
|----------|------------|
| [config.md](config.md) | `.coin/config.yaml` v2 (`goldenPath` + `version`) |
| [jenkins-setup.md](jenkins-setup.md) | Universal Jenkinsfile, K8s agents |
| [branching.md](branching.md) | Ветки, теги, версионирование |

## Компоненты monorepo

| Документ | Содержание |
|----------|------------|
| [coin-api/README.md](../coin-api/README.md) | Resolve API, scanner, migrations |
| [`coin-ui/README.md`](../coin-ui/README.md) | Admin SPA (local :8091) |
| [coin-executor/README.md](../coin-executor/README.md) | CLI runtime, CHARTER |
| [`coin-api/internal/gpcontent/seed/`](../coin-api/internal/gpcontent/seed/) | GP seed bytes (scripts, schema, orchestration) |

## Прочее

| Документ | Содержание |
|----------|------------|
| [golden-path-versioning.md](golden-path-versioning.md) | GP semver + publish model (v2) |
| [release-notes.md](release-notes.md) | QGM, smart-коммиты |

## Ключевые принципы

1. **Продукт** задаёт только `coin.goldenPath` + `coin.version` — всё остальное в manifest.
2. **coin-api** собирает manifest из PostgreSQL + git content refs.
3. **Nexus `maven-releases` / `maven-snapshots`** — runtime cache; CI работает при недоступном API.
4. **coin-executor** — stateless runtime: validate, run stages, report.
5. **Jenkinsfile.coin** — resolve → pod → executor (без Shared Library).

## Doc review (P4-04)

| Check | Status |
|-------|--------|
| Root [README.md](../README.md) — Control Plane v2 | ✅ |
| [onboarding-15min.md](how-to/onboarding-15min.md) | ✅ |
| Dead links в runbooks (wave-3 stale SQL, comms) | ✅ |
| P4-03 prod Gitea split | ⏸ corp gate |
