# `.coin/config.yaml`

Контракт между продуктовой командой и платформой Coin CI.

## Минимальный пример

```yaml
version: 1

coin:
  template: python-uv
  templateVersion: "1.0.0"
  mode: standard
  versioning:
    mode: corporate
    tagPrefix: "v"

project:
  name: my-service
  stack: python-uv

runtime:
  python: "3.13"

container:
  port: 8080
  command: "python -m my_service"

pipeline:
  test:
    enabled: true
  build:
    enabled: true
    target: container   # или package
    dockerfileTemplate: python-uv
  publish:
    enabled: true
    registry: nexus-docker   # или nexus-pypi для package
    when: tag
```

## Сборка: package vs container

- **`package`** — библиотека или pip-пакет: `uv build` → `dist/*.whl`, publish через `uv publish`. Dockerfile не нужен.
- **`container`** — микросервис: сборка контейнера по Dockerfile.
  - В компании может быть включена политика **централизованных Dockerfile**: Dockerfile‑шаблон берётся из `coin` и подкладывается в workspace при сборке.
  - В этом режиме разработчик **не ведёт Dockerfile в сервисе**, а выбирает `dockerfileTemplate` и параметры.

## Поля

| Поле | Описание |
|------|----------|
| `version` | Версия схемы (сейчас только `1`) |
| `coin.template` | Имя golden path (например `python-uv`) |
| `coin.templateVersion` | Версия шаблона (SemVer `X.Y.Z`) — нужна для политики обязательных обновлений |
| `coin.mode` | `standard` (скрипты в проекте) или `strict` (зарезервировано) |
| `coin.versioning.mode` | `corporate`: версия вычисляется Coin, а не Gradle/Maven/uv |
| `coin.versioning.tagPrefix` | Префикс релизного тега, по умолчанию `v` |
| `project.name` | Имя сервиса (отображение, Sonar) |
| `project.stack` | Стек: `python-uv`, `python-pip`, `java-maven`, `java-gradle`, `go`, `node` |
| `runtime.*` | Версия toolchain (ключ зависит от stack: `python`, `java`, …) |
| `container.port` | Порт приложения, подставляется в managed Dockerfile |
| `container.command` | Команда запуска, подставляется в managed Dockerfile |
| `pipeline.test.enabled` | Запуск тестов (по умолчанию `true`) |
| `pipeline.<stage>.preCommands` | Команды перед стандартным сценарием Coin |
| `pipeline.<stage>.commands` | Полная замена стандартного сценария stage |
| `pipeline.<stage>.postCommands` | Команды после стандартного сценария Coin |
| `pipeline.build.enabled` | Сборка артефакта |
| `pipeline.build.target` | `package` (wheel, `uv build`) или `container` (managed Dockerfile из Coin) |
| `pipeline.build.dockerfileTemplate` | (для `target: container`) имя централизованного шаблона Dockerfile (например `python-uv`, `java-spring`) |
| `pipeline.publish.enabled` | Публикация |
| `pipeline.publish.registry` | Имя credential: `nexus-pypi`, … |
| `pipeline.publish.when` | `tag` \| `branch` \| `always` \| `never` |

## Publish: `when`

- **`tag`** — только для тегов вида `v*` (или веток `v*`)
- **`branch`** — ветки `main` / `master`
- **`always`** — каждый билд
- **`never`** — отключить, даже если `enabled: true`

## Единое версионирование

Версию ведёт **Coin**, а не сборщик проекта. В проектах не настраиваем версионирование через Gradle plugins, Nebula, Maven release plugins, `uv version` и аналогичные механизмы.

Правила по умолчанию (полностью — в [docs/branching.md](branching.md)):

- tag `v1.2.3` → `COIN_VERSION=1.2.3`, image tag `1.2.3`;
- ветка `release/1.4` → `COIN_VERSION=1.4.0-rc.<build>+<shortSha>` (release candidate);
- `main` → `COIN_VERSION=0.0.0-main.<build>+<shortSha>` (snapshot);
- прочие ветки → `COIN_VERSION=0.0.0-<branch>.<build>+<shortSha>`, image tag без `+`.

Coin прокидывает:

- `COIN_VERSION` — корпоративная версия артефакта;
- `COIN_VERSION_SOURCE` — источник версии (`tag:...` или `branch:...`);
- `COIN_IMAGE_TAG` — безопасный Docker tag;
- `COIN_IMAGE_REF` — полный ref образа.

Managed Dockerfile получает `COIN_VERSION` как build arg и пишет его в `org.opencontainers.image.version`.

## Стандартные команды и расширения

По умолчанию проект **не хранит shell-скрипты**. Coin запускает стандартные сценарии из `coin-lib/resources/scripts/<stack>/`.

Пример расширения тестов:

```yaml
pipeline:
  test:
    enabled: true
    postCommands:
      - uv run ruff check .
```

Пример полной замены stage (использовать только если стандарт не подходит):

```yaml
pipeline:
  test:
    commands:
      - uv sync --frozen --all-groups
      - uv run pytest tests/integration
```

## JSON Schema

См. [`coin-lib/resources/config.schema.json`](../coin-lib/resources/config.schema.json).
