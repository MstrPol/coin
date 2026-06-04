# coin-platform

Единый репозиторий описания платформы Coin.

| Каталог | Назначение |
|---------|------------|
| `golden-paths/` | GP: profile, scripts, runtime Dockerfile, catalog policy |
| `starters/` | Скелетоны для `coin init` |
| `agents/` | CI agent images: `catalog.yaml`, `Jenkinsfile`, `stacks/` |

## Локальный стенд

```bash
cd docker
make coin-platform    # push → Gitea coin/coin-platform
make agents-build     # job agents-build в Jenkins (опционально)
```

## Переменные окружения

В CI platform клонируется в pipeline (`coinPipeline` → `.coin/platform/`):

```bash
COIN_PLATFORM_DIR=${WORKSPACE}/.coin/platform
```

Локально для coin-cli:

```bash
export COIN_PLATFORM_DIR=./coin-platform
coin platform validate
```

## Jenkins

| Job | Jenkinsfile | Регистрация |
|-----|-------------|-------------|
| `agents-build` | `agents/Jenkinsfile` | `make agents-build` |
| `demo-*` (multibranch) | `Jenkinsfile` в каждом demo-репо | `make samples` |

Service pipeline: `coinPipeline()` сам клонирует `coin/coin-platform` перед выбором agent image.
