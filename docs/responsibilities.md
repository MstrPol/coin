# Разделение ответственности

Границы владения между командами разработки и DevOps/Platform (Coin).

## Принцип

- **Проект** — код, зависимости, optional overrides pipeline.
- **Coin (platform)** — оркестрация Jenkins, agent images, golden paths, managed Dockerfile, политики версий и QG.

## Что управляет разработчик

- **Код и зависимости**: `pyproject.toml` / `pom.xml` / `go.mod` / `requirements.txt`.
- **`.coin/config.yaml`**: identity, credentials IDs, optional `jenkins.runtime`, pipeline overrides.
- **Расширения CI** через `pipeline.*.preCommands` / `postCommands` / `commands` (если разрешено политикой).
- **`container.port` / `container.command`** — параметры для managed runtime Dockerfile.

## Что управляет Platform

- **Jenkins оркестрация**: multibranch, K8s pod, credentials binding, QG.
- **CI agent images**: `coin-jenkins-agents/` — toolchain, coin CLI, docker/kaniko для pack.
- **Golden paths**: `coin-golden-paths/` — profile, scripts, runtime-only Dockerfile, catalog policy.
- **Platform CI jobs**: сборка coin-cli и agent images (Jenkinsfiles в monorepo).
- **Каталог образов**: `coin-lib/resources/images.yaml` — stack → agent image ref.
- **Версионирование**, **QG**, **security policies**, **managed Dockerfile** (генерируется в `.coin/generated/Dockerfile` при `coin run build`).

Модель сборки app: [agent-build-model.md](agent-build-model.md).

## Что не задаётся / запрещено в проекте

| Запрет | Причина |
|--------|---------|
| `Dockerfile` в репо сервиса | Managed runtime-only GP |
| `build.type`, `agent.stack` | Из `profile.yaml` golden path |
| Shell CI scripts как стандарт | GP scripts platform-owned |
| Секреты в config | Только Jenkins/Vault credential IDs |
| Своя модель версионирования | Corporate `COIN_VERSION` |

## Артефакты и владельцы

| Артефакт | Где | Владелец |
|----------|-----|----------|
| `coin-lib` | monorepo `coin` | Platform |
| `coin-cli` | monorepo `coin` | Platform |
| `coin-jenkins-agents/*` | monorepo `coin` | Platform |
| `coin-golden-paths/*` | monorepo `coin` | Platform |
| `coin-starters/*` | monorepo `coin` | Platform (эталон) |
| `coin-lib/resources/images.yaml` | monorepo `coin` | Platform |
| `.coin/config.yaml` | репо сервиса | Команда (policy-enforced) |
| App runtime Dockerfile | `.coin/generated/` в CI | Platform (render из GP) |
| App OCI image | registry | Артефакт сервиса |

> **Legacy:** `coin-lib/resources/dockerfiles/` и `resources/scripts/` — устаревшие копии, **не используются**. Канонический источник — `coin-golden-paths/`.

## Диагностика L1/L2

1. Лог: `template=`, `stack=`, `agent=`, `COIN_VERSION`.
2. `coin validate` OK → проблема platform (agent, credentials, QG) или проектных overrides.
3. `coin validate` FAIL → контракт проекта или версия GP.

## Связанные документы

- [golden-paths.md](golden-paths.md)
- [config.md](config.md)
- [agent-build-model.md](agent-build-model.md)
