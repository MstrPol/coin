# Документация Coin CI

## Начало работы

| Документ | Для кого | Содержание |
|----------|----------|------------|
| [README.md](../README.md) | все | обзор monorepo, быстрый старт |
| [jenkins-setup.md](jenkins-setup.md) | DevOps | Jenkins, platform CI, service pipelines |
| [config.md](config.md) | команды | `.coin/config.yaml` |
| [coin-starters/README.md](../coin-starters/README.md) | команды | новый репозиторий сервиса |

## Архитектура и модель сборки

| Документ | Содержание |
|----------|------------|
| [architecture.md](architecture.md) | компоненты, coin CLI, coin-lib, поток CI |
| [agent-build-model.md](agent-build-model.md) | **native build в agent + runtime-only Dockerfile** |
| [golden-paths.md](golden-paths.md) | матрица GP, profile, связь с agent |
| [golden-path-versioning.md](golden-path-versioning.md) | v1/v2, catalog.yaml, доставка GP |
| [responsibilities.md](responsibilities.md) | границы platform vs команда |
| [branching.md](branching.md) | ветки, теги, версионирование |

## Компоненты monorepo

| Документ | Содержание |
|----------|------------|
| [coin-lib/README.md](../coin-lib/README.md) | Shared Library, images.yaml |
| [coin-cli/README.md](../coin-cli/README.md) | Go CLI, команды, разработка |
| [coin-jenkins-agents/README.md](../coin-jenkins-agents/README.md) | CI agent images, platform jobs |
| [coin-golden-paths/README.md](../coin-golden-paths/README.md) | каталог GP, scripts, Dockerfile |

## Локальный стенд

| Документ | Содержание |
|----------|------------|
| [docker/README.md](../docker/README.md) | Gitea, k3s, Jenkins, bootstrap |

## Прочее

| Документ | Содержание |
|----------|------------|
| [release-notes.md](release-notes.md) | QGM, smart-коммиты |

## Ключевые принципы (кратко)

1. **coin-lib** — только оркестрация (agent, credentials). Логика — в **coin-cli**.
2. **Golden paths** — platform-owned: scripts, runtime-only Dockerfile, profile.
3. **Agent image** — CI-среда (test + native compile). Docker/kaniko — упаковка OCI.
4. **Сервисный репозиторий** — polyrepo: код + `.coin/config.yaml` + `Jenkinsfile`. Без Dockerfile.
