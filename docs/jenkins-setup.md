# Настройка Jenkins

## Global Pipeline Library

1. **Manage Jenkins** → **System** → **Global Pipeline Libraries** → **Add**.
2. Параметры:
   - **Name:** `coin-lib`
   - **Default version:** `main`
   - **Retrieval method:** Modern SCM → Git
   - **Project repository:** URL repo `coin-lib`
3. Сохранить.

Локальный стенд: `make coin-lib` → `http://gitea:3000/coin/coin-lib.git`, credential `gitea-git`.

## Kubernetes cloud

1. Плагины: **Kubernetes**, **Pipeline**, **Pipeline: Groovy**, **Git**, **Pipeline Utility Steps**, **AnsiColor**.
2. Cloud `kubernetes` + namespace для dynamic agents.
3. В сервисном `Jenkinsfile`: `coinPipeline(cloud: 'kubernetes')` при необходимости.

## Credentials

| ID | Тип | Назначение |
|----|-----|------------|
| `k3s-token` | Secret text | k3s API (bootstrap) |
| `gitea-git` | Username/Password | SCM Gitea |
| `nexus-docker` | Username/Password | Docker push/pull |
| `nexus-admin` | Username/Password | Nexus API |

Bootstrap (`casc.yaml`) создаёт creds при старте Jenkins. Mapping → env: [config.md](config.md).

## Service pipeline (репозиторий продукта)

```groovy
@Library('coin-lib') _

coinPipeline()
```

Конфигурация — `.coin/config.yaml`. Сборка — [agent-build-model.md](agent-build-model.md).

## Platform CI — agent images

| Команда | Gitea | Jenkins |
|---------|-------|---------|
| `make coin-lib` | `coin/coin-lib` | Global Library |
| `make coin-platform` | `coin/coin-platform` | push в Gitea |
| `make agents-build` | — | job **`agents-build`** |
| `make coin-cli` | `coin/coin-cli` | job **`coin-cli`** |

Job **`agents-build`**: `agents/Jenkinsfile`, сборка `ci-*` → Nexus Docker, bump `agents/catalog.yaml`.  
Подробнее — [coin-platform/README.md](../coin-platform/README.md).

## Prod-like стенд (Docker Compose)

```bash
cd docker && make bootstrap
make coin-lib
make coin-platform
make agents-build   # опционально
make coin-cli
```

| JCasC | Когда |
|-------|-------|
| `docker/jenkins/casc.yaml` | bootstrap |
| `docker/platform/jenkins/casc-coin-lib.yaml` | `make coin-lib` |
| `docker/platform/jenkins/casc-agents-build.yaml` | `make agents-build` |
| `docker/platform/jenkins/casc-coin-cli-build.yaml` | `make coin-cli` |

## Связанные документы

- [architecture.md](architecture.md)
- [agent-build-model.md](agent-build-model.md)
- [docker/README.md](../docker/README.md)
