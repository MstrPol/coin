# Настройка Jenkins

## Global Pipeline Library

1. **Manage Jenkins** → **System** → **Global Pipeline Libraries** → **Add**.
2. Параметры:
   - **Name:** `coin-lib`
   - **Default version:** `main` (или тег релиза platform)
   - **Retrieval method:** Modern SCM → Git
   - **Project repository:** URL monorepo `coin`
   - **Library path:** `coin-lib`
3. Сохранить.

На локальном стенде: `http://gitea:3000/coin/coin.git`, branch `main`, credential `gitea-git` (см. [docker/README.md](../docker/README.md)).

## Kubernetes cloud

1. Плагины: **Kubernetes**, **Pipeline**, **Pipeline: Groovy**, **Git**, **Pipeline Utility Steps**, **AnsiColor**.
2. Cloud (например `kubernetes`) + namespace для dynamic agents.
3. В сервисном `Jenkinsfile`: `coinPipeline(cloud: 'kubernetes')` при необходимости.

## Credentials

| ID | Тип | Назначение |
|----|-----|------------|
| `nexus-docker` / `coin-registry-default` | Username/Password | `COIN_REGISTRY_USER`, `COIN_REGISTRY_PASSWORD` |
| `nexus-admin` | Username/Password | publish coin-cli в Nexus (platform job) |
| `gitea-git` | Username/Password | SCM monorepo и сервисов (local) |
| `coin-publish-nexus-pypi` | Username/Password | PyPI publish (`*-lib`, roadmap) |

Mapping credentials → env: [config.md](config.md).

## Service pipeline (репозиторий продукта)

1. New Item → **Multibranch Pipeline** (или Pipeline job).
2. Branch Sources → Git, URL репозитория **сервиса** (не monorepo `coin`).
3. Script Path: `Jenkinsfile`.
4. В репозитории:

```groovy
@Library('coin-lib') _

coinPipeline()
```

Конфигурация — `.coin/config.yaml`. Сборка — [agent-build-model.md](agent-build-model.md).

## Platform CI (monorepo `coin`)

Platform jobs собирают **coin-cli** и **agent images**. Два job.

| Job | SCM | Jenkinsfile | Результат |
|-----|-----|-------------|-----------|
| `coin-cli` | `coin/coin.git` | `coin-cli/Jenkinsfile` | `coin_linux_<arch>` → Nexus |
| `coin-agents` | `coin/coin.git` | `coin-jenkins-agents/Jenkinsfile` | `ci-*:{runtime}-r{N}` → registry |

Параметры `coin-agents`: Active Choices `STACK` → `RUNTIME` (из `catalog.yaml`), `BUILD_ALL`.

Agent build: `docker build` (monorepo root context) → push registry → bump `catalog.yaml`.  
`coin` CLI в образ не зашивается.  
Подробнее — [coin-jenkins-agents/README.md](../coin-jenkins-agents/README.md).

### Promote agent images → images.yaml

После сборки (или вручную на local стенде):

```bash
docker/scripts/sync-agent-images.sh
```

Обновляет `coin-lib/resources/images.yaml` из `catalog.yaml`:

```yaml
stacks:
  python-uv:
    "3.13":
      image: coin/ci-python-uv:3.13-r1
      digest: "sha256:…"
      rev: 1
```

В prod-like стенде предпочтительно `make platform-build` в [docker/](../docker/README.md).

## Prod-like стенд (Docker Compose)

```bash
cd docker && make bootstrap
```

Поднимает Gitea, Nexus, registry, k3s, Jenkins; пушит `coin` + demo; запускает platform build.

## Связанные документы

- [architecture.md](architecture.md)
- [agent-build-model.md](agent-build-model.md)
- [docker/README.md](../docker/README.md)
