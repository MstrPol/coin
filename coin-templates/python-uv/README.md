# Python + uv (Coin template)

Скопируйте содержимое этой папки в корень репозитория сервиса.

## Во что собирается проект?

Зависит от `pipeline.build.target` в `.coin/config.yaml`:

| `target` | Результат | Кто ведёт артефакт |
|----------|-----------|-------------------|
| **`package`** | Python wheel/sdist в `dist/` → publish в Nexus/PyPI (`uv publish`) | `pyproject.toml` — команда; скрипты CI — платформа (`coin-lib`) |
| **`container`** | Docker-образ приложения → push в container registry | **Managed Dockerfile из Coin** + параметры из `.coin/config.yaml` |

### Два разных «образа» (не путать)

1. **`coin/ci-python-uv`** (`coin-images/`) — среда **CI** (Jenkins agent). Ведёт **platform**.
2. **Managed Dockerfile** (`coin-lib/resources/dockerfiles/python-uv/Dockerfile`) — runtime для деплоя в K8s. Ведёт **platform**, проект передаёт только параметры (`container.port`, `container.command`).

## Dockerfile

В репозитории сервиса **не должно быть Dockerfile** и **не нужен `.dockerignore`** для CI.

Coin перед сборкой генерирует `.coin/generated/Dockerfile` из централизованного шаблона, managed `.dockerignore` в workspace и передаёт путь в `COIN_DOCKERFILE`.

Параметры:

```yaml
container:
  port: 8080
  command: "python -m my_service"
```

Если сервису нужны системные пакеты или нестандартный runtime — это не правка Dockerfile в проекте, а запрос на новый managed Dockerfile template в Coin.

## Где живут команды

| Что | Где |
|-----|-----|
| **Когда** запускать stage | `.coin/config.yaml` + `coinPipeline` |
| **Что** выполнять по умолчанию | `coin-lib/resources/scripts/python-uv/*.sh` |
| Расширения/переопределения | `pipeline.<stage>.preCommands/postCommands/commands` в `.coin/config.yaml` |
| Образ CI, pod, секреты | `coin-lib` + `coin-images` |

Стандартный путь не требует `.coin/scripts/*.sh` в проекте. Если нужно расширить тесты:

```yaml
pipeline:
  test:
    postCommands:
      - uv run ruff check .
```

Если стандарт полностью не подходит, можно заменить stage:

```yaml
pipeline:
  test:
    commands:
      - uv sync --frozen --all-groups
      - uv run pytest tests/integration
```

## Требования

- Jenkins + Global Pipeline Library `coin-lib`
- K8s cloud + `coin/ci-python-uv:3.13`
- Для **package**: credential `coin-publish-nexus-pypi`, `COIN_NEXUS_PYPI_URL`
- Для **container**: Docker на агенте или Kaniko (`ci-tooling`), `COIN_REGISTRY_PREFIX`, credential `coin-publish-nexus-docker`

## Локально

```bash
uv sync --all-groups
uv run pytest

# package
uv build

# container локально можно проверить только через Coin/Jenkins,
# потому что Dockerfile генерируется платформой.
```
